package builder

type Builder interface {
	// New(threads int, dryRun bool) *Builder
	Init() error
	Build(dockerfile, imageName string, labels map[string]string, contextDir string, verbose bool)
	Tag(imageName, taggedImage string, verbose bool)
	Push(taggedImage string, verbose bool)
	Remove(imageName string, verbose bool)
	Run(stage Stage) error
	SetThreads(threads int)
	SetDryRun(dryRun bool)
	Shutdown() error
}

// func DefaultRun(b Builder, stage Stage) error {
//     switch stage {
//     case Build:
//         return b.Build()
//     case Tag:
//         return b.Tag()
//     case Push:
//         return b.Push()
//     case Remove:
//         return b.Remove()
//     default:
//         return fmt.Errorf("unknown stage: %s", stage)
//     }
// }
