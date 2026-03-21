package builder

import (
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/parser"
)

type Engine interface {
	ExecutePlan(plan *parser.Plan, flags *config.Flags) error
}
