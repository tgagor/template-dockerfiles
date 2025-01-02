package builder

import (
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
)

type Builder interface {
	// New(threads int, dryRun bool) *Builder
	Init() error
	SetFlags(flags *config.Flags)
	Build(image *image.Image)
	Squash(image *image.Image)
	// Tag(image *image.Image, flags *config.Flags)
	// Push(image *image.Image, flags *config.Flags)
	// Remove(image *image.Image, flags *config.Flags)
	setThreads(threads int)
	setDryRun(dryRun bool)
	Shutdown() error

	Run() error
	// RunBuilding() error
	// RunSquashing() error
	// RunTagging() error
	// RunPushing() error
	// RunCleanup() error
}
