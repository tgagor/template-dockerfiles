package builder

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
	"github.com/tgagor/template-dockerfiles/pkg/runner"
)

type BuildxBuilder struct {
	flags  *config.Flags
	images []*image.Image

	buildTasks   *runner.Runner
	taggingTasks *runner.Runner
	cleanupTasks *runner.Runner
}

func (b *BuildxBuilder) Init() error {
	// create
	// docker buildx create \
	//   --name multi-platform-builder \
	//   --driver docker-container \
	//   --use
	b.buildTasks = runner.New()
	b.taggingTasks = runner.New()
	b.cleanupTasks = runner.New()

	log.Info().Str("engine", "buildx").Msg("Initializing")
	return nil
}

func (b *BuildxBuilder) SetFlags(flags *config.Flags) {
	b.flags = flags
	b.SetThreads(flags.Threads)
	b.SetDryRun(!flags.Build)
}

func (b *BuildxBuilder) SetThreads(threads int) {
	b.buildTasks.Threads(threads)
	// b.tagTasks have to use 1 thread
	b.cleanupTasks.Threads(threads)
}

func (b *BuildxBuilder) SetDryRun(dryRun bool) {
	b.buildTasks.DryRun(dryRun)
	b.taggingTasks.DryRun(dryRun)
	b.cleanupTasks.DryRun(dryRun)
}

func (b *BuildxBuilder) Queue(image *image.Image) {
	b.images = append(b.images, image)
}

func (b *BuildxBuilder) Build(img *image.Image) {
	builder := cmd.New("docker").Arg("buildx").Arg("build")
	if len(img.Platforms) > 0 {
		builder.Arg(platformsToArgs(img.Platforms)...)
	}
	builder.Arg("-f", img.Dockerfile).
		Arg("-t", img.UniqName()).
		Arg(labelsToArgs(img.Labels)...).
		Arg(buildArgsToArgs(img.BuildArgs)...).
		Arg("--load").
		Arg(img.BuildContextDir).SetVerbose(b.flags.Verbose)
	b.buildTasks.AddTask(builder)
	b.Remove(img.UniqName()) // this image is temporary, remove it after build
}

func (b *BuildxBuilder) Squash(img *image.Image) {}

func (b *BuildxBuilder) TagAndPush(img *image.Image) {
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

	// let push here
	if b.flags.Push {
		tagger.Arg("--push")
		tagger.PreInfo("Tagging and pushing " + img.UniqName() + " with tags: " + strings.Join(img.Tags(), ", "))
	}

	// builder.Arg("--output", "type=image,name=" + imageName) // required for multi-platform builds
	tagger.Arg(img.BuildContextDir).
		SetVerbose(b.flags.Verbose)

	if !b.flags.Push {
		tagger.PreInfo("Tagging " + img.UniqName() + " with tags: " + strings.Join(img.Tags(), ", "))
	}
	b.taggingTasks.AddTask(tagger)
}

func (b *BuildxBuilder) Remove(imageName string) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(imageName).
		SetVerbose(b.flags.Verbose)
	b.cleanupTasks.AddTask(remover)
}

func (b *BuildxBuilder) Run() error {
	// build images in parallel to fill the cache
	if b.flags.Build {
		for _, img := range b.images {
			b.Build(img)
		}

		if err := b.buildTasks.Run(); err != nil {
			return err
		}
	} else {
		log.Warn().Msg("Skipping building images. Use --build flag to build images.")
		return nil
	}

	if b.flags.Build && b.flags.Squash {
		log.Warn().Msg("Squash is not supported for buildx engine. Skipping.")
	}

	// tag and push images based on the cache
	if b.flags.Build || b.flags.Push {
		for _, img := range b.images {
			b.TagAndPush(img)
		}

		if err := b.taggingTasks.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (b *BuildxBuilder) Terminate() error {
	if b.flags.Build {
		if err := b.cleanupTasks.Run(); err != nil {
			return err
		}
	}

	// Cleanup temporary dockerfiles
	if b.flags.Delete {
		for _, img := range b.images {
			defer img.RemoveTemporaryDockerfile()
		}
	}

	return nil
}
