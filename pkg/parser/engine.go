package parser

import "github.com/tgagor/template-dockerfiles/pkg/config"

type Engine interface {
	Parse(cfg *config.Config, flags *config.Flags) error
	// Parse(p *ParseStrategy)
}
