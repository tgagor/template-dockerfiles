package parser

import "github.com/tgagor/template-dockerfiles/pkg/config"

type Engine interface {
	ExecutePlan(plan *Plan, flags *config.Flags) error
}
