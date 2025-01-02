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
	flags        *config.Flags
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
	b.setThreads(flags.Threads)
	b.setDryRun(!flags.Build)
}

func (b *BuildxBuilder) setThreads(threads int) {
	b.buildTasks.Threads(threads)
	// b.tagTasks have to use 1 thread
	b.cleanupTasks.Threads(threads)
}

func (b *BuildxBuilder) setDryRun(dryRun bool) {
	b.buildTasks.DryRun(dryRun)
	b.taggingTasks.DryRun(dryRun)
	b.cleanupTasks.DryRun(dryRun)
}

func (b *BuildxBuilder) Build(img *image.Image) {
	builder := cmd.New("docker").Arg("buildx").Arg("build")
	if len(img.Platforms) > 0 {
		builder.Arg(platformsToArgs(img.Platforms)...)
	}
	builder.Arg("-f", img.Dockerfile).
		Arg("-t", img.Name).
		Arg(labelsToArgs(img.Labels)...).
		Arg(buildArgsToArgs(img.BuildArgs)...).
		Arg("--load").
		Arg(img.BuildContextDir).SetVerbose(b.flags.Verbose)
	b.buildTasks.AddTask(builder)
	b.Remove(img) // this image is temporary, remove it after build
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
		tagger.PreInfo("Tagging and pushing " + img.Name + " with tags: " + strings.Join(img.Tags(), ", "))
	}

	// builder.Arg("--output", "type=image,name=" + imageName) // required for multi-platform builds
	tagger.Arg(img.BuildContextDir).
		SetVerbose(b.flags.Verbose).
		PreInfo("Tagging " + img.Name + " with tags: " + strings.Join(img.Tags(), ", "))
	b.taggingTasks.AddTask(tagger)
}


func (b *BuildxBuilder) Remove(img *image.Image) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(img.Name).
		SetVerbose(b.flags.Verbose)
	b.cleanupTasks.AddTask(remover)
}

func (b *BuildxBuilder) Run() error {
	if b.flags.Build {
		if err := b.buildTasks.Run(); err != nil {
			return err
		}
	}

	if b.flags.Build && b.flags.Squash {
		log.Warn().Msg("Squash is not supported for buildx engine. Skipping.")
	}

	if b.flags.Build || b.flags.Push {
		if err := b.taggingTasks.Run(); err != nil {
			return err
		}
	}

	if b.flags.Build {
		if err := b.cleanupTasks.Run(); err != nil {
			return err
		}
	}
	return nil
}

// func (b *BuildxBuilder) runBuilding() error {
// 	if b.flags.Build {
// 		return b.buildTasks.Run()
// 	} else {
// 		log.Warn().Msg("Skipping building images")
// 		return nil
// 	}
// }

func (b *BuildxBuilder) Shutdown() error {
	return nil
}
