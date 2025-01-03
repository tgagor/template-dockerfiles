package builder

import (
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
)

type Builder interface {
	// initialize builder if needed
	Init() error
	SetFlags(flags *config.Flags)

	// collect images for building
	Queue(image *image.Image)

	// execute the build process
	Run() error

	// cleanup tasks
	Terminate() error
}
