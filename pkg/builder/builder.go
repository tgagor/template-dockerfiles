package builder

import (
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
)

type Builder interface {
	// New(threads int, dryRun bool) *Builder
	Init() error
	Build(image *image.Image, flags *config.Flags)
	Squash(image *image.Image, flags *config.Flags)
	Tag(image *image.Image, flags *config.Flags)
	Push(image *image.Image, flags *config.Flags)
	Remove(image *image.Image, flags *config.Flags)
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
