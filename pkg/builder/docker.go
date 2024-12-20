package builder

import (
	"fmt"

	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/runner"
)

type DockerBuilder struct {
	buildTasks   runner.Runner
	tagTasks     runner.Runner
	pushTasks    runner.Runner
	cleanupTasks runner.Runner
}

func (b *DockerBuilder) Init() error {
	return nil
}

func (b *DockerBuilder) SetThreads(threads int) {
	b.buildTasks = b.buildTasks.Threads(threads)
	// b.tagTasks have to use 1 thread
	b.pushTasks = b.pushTasks.Threads(threads)
	b.cleanupTasks = b.cleanupTasks.Threads(threads)
}

func (b *DockerBuilder) SetDryRun(dryRun bool) {
	b.buildTasks = b.buildTasks.DryRun(dryRun)
	b.tagTasks = b.tagTasks.DryRun(dryRun)
	b.pushTasks = b.pushTasks.DryRun(dryRun)
	b.cleanupTasks = b.cleanupTasks.DryRun(dryRun)
}

func (b *DockerBuilder) Build(dockerfile, imageName string, labels map[string]string, contextDir string, verbose bool) {
	builder := cmd.New("docker").Arg("build").
		Arg("-f", dockerfile).
		Arg("-t", imageName).
		Arg(labelsToArgs(labels)...).
		Arg(contextDir).
		SetVerbose(verbose)
	b.buildTasks = b.buildTasks.AddTask(builder)
}

func (b *DockerBuilder) Tag(imageName, taggedImage string, verbose bool) {
	tagger := cmd.New("docker").Arg("tag").
		Arg(imageName).
		Arg(taggedImage).
		SetVerbose(verbose).
		PreInfo("Tagging " + taggedImage)
	b.tagTasks = b.tagTasks.AddTask(tagger)
}

func (b *DockerBuilder) Push(taggedImage string, verbose bool) {
	pusher := cmd.New("docker").Arg("push").
		Arg(taggedImage).
		PreInfo("Pushing " + taggedImage)
	if !verbose {
		pusher = pusher.Arg("--quiet")
	}
	b.pushTasks = b.pushTasks.AddTask(pusher)
}

func (b *DockerBuilder) Remove(imageName string, verbose bool) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(imageName).
		SetVerbose(verbose)
	b.cleanupTasks = b.cleanupTasks.AddTask(remover)
}

func (b *DockerBuilder) Run(stage Stage) error {
	switch stage {
	case Build:
		return b.buildTasks.Run()
	case Tag:
		return b.tagTasks.Run()
	case Push:
		return b.pushTasks.Run()
	case Remove:
		return b.cleanupTasks.Run()
	default:
		return fmt.Errorf("unknown stage: %s", stage)
	}
}

func labelsToArgs(labels map[string]string) []string {
	args := []string{}
	for k, v := range labels {
		args = append(args, "--label", k+"="+v)
	}
	return args
}
