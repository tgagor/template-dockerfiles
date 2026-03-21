package builder

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
)

type BuildxBuilder struct {
	flags *config.Flags
}

func (b *BuildxBuilder) Init() error {
	log.Info().Str("engine", "buildx").Msg("Initializing")
	return nil
}

func (b *BuildxBuilder) SetFlags(flags *config.Flags) {
	b.flags = flags
}

func (b *BuildxBuilder) Process(ctx context.Context, img *image.Image) error {
	if b.flags.Build {
		builder := cmd.New("docker").Arg("buildx").Arg("build")
		if len(img.Platforms) > 0 {
			builder.Arg(platformsToArgs(img.Platforms)...)
		}
		builder.Arg(img.Options...).
			Arg("-f", img.Dockerfile).
			Arg("-t", img.UniqName()).
			Arg(labelsToArgs(img.Labels)...).
			Arg(buildArgsToArgs(img.BuildArgs)...).
			Arg("--load").
			Arg(img.BuildContextDir).
			PreInfo("Building " + img.UniqName()).
			PostInfo("Built " + img.UniqName()).
			SetVerbose(b.flags.Verbose)
		if _, err := builder.Run(ctx); err != nil {
			return err
		}
		defer b.Remove(ctx, img.UniqName())
	}

	if b.flags.Build && b.flags.Squash {
		log.Warn().Msg("Squash is not supported for buildx engine. Skipping.")
	}

	if b.flags.Build || b.flags.Push {
		tagger := cmd.New("docker").Arg("buildx").Arg("build")
		if len(img.Platforms) > 0 {
			tagger.Arg(platformsToArgs(img.Platforms)...)
		}
		tagger.Arg("-f", img.Dockerfile).
			Arg(labelsToArgs(img.Labels)...).
			Arg(buildArgsToArgs(img.BuildArgs)...).
			Arg("--load")

		// collect tagging commands to keep order
		for _, imageTag := range img.Tags() {
			tagger.Arg("-t", imageTag)
		}

		if b.flags.Push {
			tagger.Arg("--push")
			tagger.PreInfo("Tagging and pushing " + img.UniqName() + " with tags: " + strings.Join(img.Tags(), ", "))
		}

		tagger.Arg(img.BuildContextDir).SetVerbose(b.flags.Verbose)

		if !b.flags.Push {
			tagger.PreInfo("Tagging " + img.UniqName() + " with tags: " + strings.Join(img.Tags(), ", "))
		}

		if _, err := tagger.Run(ctx); err != nil {
			return err
		}
	}

	if b.flags.Delete {
		img.RemoveTemporaryDockerfile()
	}

	return nil
}

func (b *BuildxBuilder) Remove(ctx context.Context, imageName string) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(imageName).
		SetVerbose(b.flags.Verbose)
	_, _ = remover.Run(ctx)
}

func (b *BuildxBuilder) Terminate() error {
	return nil
}
