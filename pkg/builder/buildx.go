package builder

import (
	"fmt"

	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/runner"
)

type BuildxBuilder struct {
	buildTasks   *runner.Runner
	tagTasks     *runner.Runner
	pushTasks    *runner.Runner
	cleanupTasks *runner.Runner
}

func (b *BuildxBuilder) Init() error {
	// create
	// docker buildx create \
	//   --name multi-platform-builder \
	//   --driver docker-container \
	//   --use
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
	b.tagTasks.DryRun(dryRun)
	b.pushTasks.DryRun(dryRun)
	b.cleanupTasks.DryRun(dryRun)
}

func (b *BuildxBuilder) Build(dockerfile, imageName string, labels map[string]string, contextDir string, verbose bool) {
	builder := cmd.New("docker").Arg("buildx").Arg("build").
		Arg("-f", dockerfile).
		Arg("-t", imageName).
		Arg(labelsToArgs(labels)...).
		Arg(contextDir).
		SetVerbose(verbose)
	b.buildTasks.AddTask(builder)
}

func (b *BuildxBuilder) Tag(imageName, taggedImage string, verbose bool) {
	tagger := cmd.New("docker").Arg("tag").
		Arg(imageName).
		Arg(taggedImage).
		SetVerbose(verbose).
		PreInfo("Tagging " + taggedImage)
	b.tagTasks.AddUniq(tagger)
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

func (b *BuildxBuilder) Run(stage Stage) error {
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

func (b *BuildxBuilder) Shutdown() error {
	return nil
}

// func labelsToArgs(labels map[string]string) []string {
// 	args := []string{}
// 	for k, v := range labels {
// 		args = append(args, "--label", k+"="+v)
// 	}
// 	return args
// }
