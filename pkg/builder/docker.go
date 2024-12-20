package builder

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/runner"
)

type DockerBuilder struct {
	buildTasks   *runner.Runner
	taggingTasks *runner.Runner
	pushTasks    *runner.Runner
	cleanupTasks *runner.Runner
}

func (b *DockerBuilder) Init() error {
	b.buildTasks = runner.New()
	b.taggingTasks = runner.New()
	b.pushTasks = runner.New()
	b.cleanupTasks = runner.New()
	return nil
}

func (b *DockerBuilder) SetThreads(threads int) {
	b.buildTasks.Threads(threads)
	// b.tagTasks have to use 1 thread
	b.pushTasks.Threads(threads)
	b.cleanupTasks.Threads(threads)
}

func (b *DockerBuilder) SetDryRun(dryRun bool) {
	b.buildTasks.DryRun(dryRun)
	b.taggingTasks.DryRun(dryRun)
	b.pushTasks.DryRun(dryRun)
	b.cleanupTasks.DryRun(dryRun)
}

func (b *DockerBuilder) Build(dockerfile, imageName string, labels map[string]string, contextDir string, verbose bool) {
	builder := cmd.New("docker").Arg("build").
		Arg("-f", dockerfile).
		Arg("-t", imageName).
		Arg(labelsToArgs(labels)...).
		Arg(contextDir).
		PreInfo("Building " + imageName).
		SetVerbose(verbose)
	b.buildTasks.AddTask(builder)
}

func (b *DockerBuilder) Tag(imageName, taggedImage string, verbose bool) {
	tagger := cmd.New("docker").Arg("tag").
		Arg(imageName).
		Arg(taggedImage).
		PreInfo("Tagging " + taggedImage).
		SetVerbose(verbose)
	b.taggingTasks.AddUniq(tagger)
}

func (b *DockerBuilder) Push(taggedImage string, verbose bool) {
	pusher := cmd.New("docker").Arg("push").
		Arg(taggedImage).
		PreInfo("Pushing " + taggedImage)
	if !verbose {
		pusher.Arg("--quiet")
	}
	b.pushTasks.AddTask(pusher)
}

func (b *DockerBuilder) Remove(imageName string, verbose bool) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(imageName).
		SetVerbose(verbose)
	b.cleanupTasks.AddTask(remover)
}

func (b *DockerBuilder) Run(stage Stage) error {
	log.Debug().Str("stage", stage.String()).Msg("Running stage: ")
	switch stage {
	case Build:
		log.Info().Msg("Running build stage")
		return b.buildTasks.Run()
	case Tag:
		log.Info().Msg("Running tagging stage")
		log.Info().Int("tag tasks", b.taggingTasks.CountTasks()).Msg("Tagging tasks")
		return b.taggingTasks.Run()
	case Push:
		log.Info().Msg("Running push stage")
		return b.pushTasks.Run()
	case Remove:
		log.Info().Msg("Running cleanup stage")
		return b.cleanupTasks.Run()
	default:
		return fmt.Errorf("unknown stage: %s", stage)
	}
}

func (b *DockerBuilder) Shutdown() error {
	return nil
}

func labelsToArgs(labels map[string]string) []string {
	args := []string{}
	for k, v := range labels {
		args = append(args, "--label", k+"="+v)
	}
	return args
}
