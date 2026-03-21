package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

type DockerBuilder struct {
	flags *config.Flags
}

func (b *DockerBuilder) Init() error {
	log.Info().Str("engine", "docker").Msg("Initializing")
	return nil
}

func (b *DockerBuilder) SetFlags(flags *config.Flags) {
	b.flags = flags
}

func (b *DockerBuilder) Process(ctx context.Context, img *image.Image) error {
	if b.flags.Build {
		builder := cmd.New("docker").Arg("build").
			Arg(img.Options...).
			Arg("-f", img.Dockerfile).
			Arg("-t", img.UniqName()).
			Arg(labelsToArgs(img.Labels)...).
			Arg(buildArgsToArgs(img.BuildArgs)...).
			Arg(img.BuildContextDir).
			PreInfo("Building " + img.UniqName()).
			PostInfo("Built " + img.UniqName()).
			SetVerbose(b.flags.Verbose)
		if _, err := builder.Run(ctx); err != nil {
			return err
		}
		// ensure cleanup of the transient image
		defer b.Remove(ctx, img.UniqName())
	}

	if b.flags.Build && b.flags.Squash {
		if err := b.Squash(ctx, img); err != nil {
			return err
		}
	}

	if b.flags.Build {
		for _, tag := range img.Tags() {
			tagger := cmd.New("docker").Arg("tag").
				Arg(img.UniqName()).
				Arg(tag).
				PreInfo("Tagging " + tag).
				SetVerbose(b.flags.Verbose)
			if _, err := tagger.Run(ctx); err != nil {
				return err
			}
		}
	}

	if b.flags.Push {
		for _, tag := range img.Tags() {
			pusher := cmd.New("docker").Arg("push").
				Arg(tag).
				PreInfo("Pushing " + tag)
			if !b.flags.Verbose {
				pusher.Arg("--quiet")
			}
			if _, err := pusher.Run(ctx); err != nil {
				return err
			}
		}
	}

	if b.flags.Delete {
		img.RemoveTemporaryDockerfile()
	}

	return nil
}

func (b *DockerBuilder) Squash(ctx context.Context, img *image.Image) error {
	containerName := "run-" + util.SanitizeForFileName(img.UniqName())

	runItFirst := cmd.New("docker").Arg("run").
		Arg("--name", containerName).
		Arg(img.UniqName()).
		Arg("true").
		SetVerbose(b.flags.Verbose)
	if _, err := runItFirst.Run(ctx); err != nil {
		return err
	}

	imgMetadata, err := InspectImage(img.UniqName())
	if err != nil {
		return fmt.Errorf("couldn't inspect Docker image %s: %w", img.UniqName(), err)
	}
	log.Trace().Interface("data", imgMetadata).Msg("Docker inspect result")
	sizeBefore := imgMetadata[0].Size

	tmpTarFile := containerName + ".tar"
	exportIt := cmd.New("docker").Arg("export").
		Arg(containerName).
		Arg("-o", tmpTarFile).
		PreInfo(fmt.Sprintf("Squashing %s", img.UniqName())).
		SetVerbose(b.flags.Verbose)
	if _, err := exportIt.Run(ctx); err != nil {
		return err
	}
	defer util.RemoveFile(tmpTarFile)

	cleanupCmd := cmd.New("docker").Arg("rm").Arg("-f").Arg(containerName)
	if _, err := cleanupCmd.Run(ctx); err != nil {
		return err
	}

	importIt := cmd.New("docker").Arg("import")
	for _, item := range imgMetadata {
		// paring ENV
		for _, env := range item.Config.Env {
			importIt.Arg("--change", "ENV "+env)
		}

		// parsing CMD
		if command, err := json.Marshal(item.Config.Cmd); err != nil {
			log.Error().Err(err).Str("image", img.UniqName()).Msg("Can't parse CMD")
		} else {
			importIt.Arg("--change", "CMD "+string(command))
		}

		// parsing VOLUME
		if vol, err := json.Marshal(item.Config.Volumes); err != nil {
			log.Error().Err(err).Str("image", img.UniqName()).Msg("Can't parse VOLUME")
		} else {
			importIt.Arg("--change", "VOLUME "+string(vol))
		}

		// parsing LABELS
		for key, value := range item.Config.Labels {
			importIt.Arg("--change", fmt.Sprintf("LABEL %s=\"%s\"", key, strings.ReplaceAll(value, "\n", "")))
		}

		// parsing ENTRYPOINT
		if entrypoint, err := json.Marshal(item.Config.Entrypoint); err != nil {
			log.Error().Err(err).Str("image", img.UniqName()).Msg("Can't parse ENTRYPOINT")
		} else {
			importIt.Arg("--change", "CMD "+string(entrypoint))
		}

		// parsing WORKDIR
		if item.Config.WorkingDir != "" {
			importIt.Arg("--change", "WORKDIR "+item.Config.WorkingDir)
		}
	}
	importIt.Arg(tmpTarFile).Arg(img.UniqName()).SetVerbose(b.flags.Verbose)
	if _, err := importIt.Run(ctx); err != nil {
		return err
	}

	// remove interim images
	oldImgHash := strings.TrimPrefix(imgMetadata[0].Id, "sha256:")[:12]
	b.Remove(ctx, oldImgHash)

	// Log reduction
	imgMetadataAfter, err := InspectImage(img.UniqName())
	if err == nil {
		sizeAfter := imgMetadataAfter[0].Size
		percentage := float64(sizeAfter)*100/float64(sizeBefore) - 100
		log.Info().Str("image", img.UniqName()).Str("was", util.ByteCountIEC(sizeBefore)).Str("is", util.ByteCountIEC(sizeAfter)).Str("reduction", fmt.Sprintf("%.1f%%", percentage)).Msg("Squashed")
	}

	return nil
}

func (b *DockerBuilder) Remove(ctx context.Context, imageName string) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(imageName).
		SetVerbose(b.flags.Verbose)
	// fire and forget for cleanup
	_, _ = remover.Run(ctx)
}

func (b *DockerBuilder) Terminate() error {
	return nil
}
