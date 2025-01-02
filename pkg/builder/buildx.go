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
	buildTasks   *runner.Runner
	taggingTasks *runner.Runner
	pushTasks    *runner.Runner
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
	b.pushTasks = runner.New()
	b.cleanupTasks = runner.New()

	log.Info().Str("engine", "buildx").Msg("Initializing")
	return nil
}

func (b *BuildxBuilder) SetThreads(threads int) {
	b.buildTasks.Threads(threads)
	// b.tagTasks have to use 1 thread
	b.pushTasks.Threads(threads)
	b.cleanupTasks.Threads(threads)
}

func (b *BuildxBuilder) SetDryRun(dryRun bool) {
	b.buildTasks.DryRun(dryRun)
	b.taggingTasks.DryRun(dryRun)
	b.pushTasks.DryRun(dryRun)
	b.cleanupTasks.DryRun(dryRun)
}

func (b *BuildxBuilder) Build(img *image.Image, flags *config.Flags) {
	builder := cmd.New("docker").Arg("buildx").Arg("build")
	if len(img.Platforms) > 0 {
		builder.Arg(platformsToArgs(img.Platforms)...)
	}
	builder.Arg("-f", img.Dockerfile).
		Arg("-t", img.Name).
		Arg(labelsToArgs(img.Labels)...).
		Arg(buildArgsToArgs(img.BuildArgs)...).
		Arg("--load").
		Arg(img.BuildContextDir).SetVerbose(flags.Verbose)
	b.buildTasks.AddTask(builder)
}

func (b *BuildxBuilder) Squash(img *image.Image, flags *config.Flags) {}

func (b *BuildxBuilder) Tag(img *image.Image, flags *config.Flags) {
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
	if flags.Push {
		tagger.Arg("--push")
	}

	// builder.Arg("--output", "type=image,name=" + imageName) // required for multi-platform builds
	tagger.Arg(img.BuildContextDir).
		SetVerbose(flags.Verbose).
		PreInfo("Tagging " + img.Name + " with tags: " + strings.Join(img.Tags(), ", "))
	b.taggingTasks.AddTask(tagger)
}

func (b *BuildxBuilder) Push(img *image.Image, flags *config.Flags) {
	pusher := cmd.New("docker").Arg("buildx").Arg("build")
	if len(img.Platforms) > 0 {
		pusher.Arg(platformsToArgs(img.Platforms)...)
	}
	pusher.Arg("-f", img.Dockerfile).
		Arg(labelsToArgs(img.Labels)...).
		Arg(buildArgsToArgs(img.BuildArgs)...).
		Arg("--load")

	// collect tagging commands to keep order
	for _, imageTag := range img.Tags() {
		pusher.Arg("-t", imageTag)
	}

	// let push here
	if flags.Push {
		pusher.Arg("--push")
	}

	// builder.Arg("--output", "type=image,name=" + imageName) // required for multi-platform builds
	pusher.Arg(img.BuildContextDir).
		SetVerbose(flags.Verbose).
		PreInfo("Pushing " + img.Name + " with tags: " + strings.Join(img.Tags(), ", "))
	b.taggingTasks.AddTask(pusher)
}

func (b *BuildxBuilder) Remove(img *image.Image, flags *config.Flags) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(img.Name).
		SetVerbose(flags.Verbose)
	b.cleanupTasks.AddTask(remover)
}

func (b *BuildxBuilder) RunBuilding() error {
	return b.buildTasks.Run()
}
func (b *BuildxBuilder) RunSquashing() error {
	log.Warn().Msg("Squash is not supported for buildx")
	return nil
}
func (b *BuildxBuilder) RunTagging() error {
	return b.taggingTasks.Run()
}
func (b *BuildxBuilder) RunPushing() error {
	return b.pushTasks.Run()
}
func (b *BuildxBuilder) RunCleanup() error {
	return b.cleanupTasks.Run()
}

func (b *BuildxBuilder) Shutdown() error {
	return nil
}
