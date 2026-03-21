package builder

import (
	"context"

	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
)

type Builder interface {
	// initialize builder if needed
	Init() error
	SetFlags(flags *config.Flags)

	// Process handles the sequential build, tag, and push flow for a single image
	Process(ctx context.Context, img *image.Image) error

	// cleanup tasks
	Terminate() error
}
