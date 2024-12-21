package builder

import (
	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/cmd"
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

func (b *BuildxBuilder) Build(dockerfile, imageName string, configSet map[string]interface{}, contextDir string, verbose bool) {
	platforms := configSet["platforms"].([]string)
	labels := configSet["labels"].(map[string]string)
	builder := cmd.New("docker").Arg("buildx").Arg("build")
	if len(platforms) > 0 {
		builder.Arg(platformsToArgs(platforms)...)
	}
	builder.Arg("-f", dockerfile).Arg("-t", imageName).Arg(labelsToArgs(labels)...).Arg(contextDir).SetVerbose(verbose)
	b.buildTasks.AddTask(builder)
}

func (b *BuildxBuilder) Squash(imageName string, verbose bool) {}

func (b *BuildxBuilder) Tag(imageName, taggedImage string, verbose bool) {
	tagger := cmd.New("docker").Arg("tag").
		Arg(imageName).
		Arg(taggedImage).
		SetVerbose(verbose).
		PreInfo("Tagging " + taggedImage)
	b.taggingTasks.AddUniq(tagger)
}

func (b *BuildxBuilder) Push(taggedImage string, verbose bool) {
	pusher := cmd.New("docker").Arg("push").
		Arg(taggedImage)
	if !verbose {
		pusher.Arg("--quiet")
	}
	b.pushTasks.AddTask(pusher)
}

func (b *BuildxBuilder) Remove(imageName string, verbose bool) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(imageName).
		SetVerbose(verbose)
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
