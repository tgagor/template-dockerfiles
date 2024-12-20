package builder

type Builder interface {
	// New(threads int, dryRun bool) *Builder
	Init() error
	Build(dockerfile, imageName string, labels map[string]string, contextDir string, verbose bool)
	Squash(imageName string, verbose bool)
	Tag(imageName, taggedImage string, verbose bool)
	Push(taggedImage string, verbose bool)
	Remove(imageName string, verbose bool)
	// Run(stage Stage) error
	SetThreads(threads int)
	SetDryRun(dryRun bool)
	Shutdown() error

	RunBuilding() error
	RunSquashing() error
	RunTagging() error
	RunPushing() error
	RunCleanup() error
}
