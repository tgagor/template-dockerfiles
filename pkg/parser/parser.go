package parser

import "github.com/tgagor/template-dockerfiles/pkg/config"

type Parser struct {
	engine Engine
	cfg    *config.Config
	flags  *config.Flags
}

func NewParser(cfg *config.Config, flags *config.Flags) *Parser {
	var e Engine
	switch flags.Engine {
	case "buildx":
		e = &BuildxEngine{}
	// case "kaniko":
	// 	buildEngine = &builder.KanikoBuilder{}
	// case "podman":
	// 	buildEngine = &builder.PodmanBuilder{}
	default:
		e = &DockerEngine{}
	}
	return &Parser{
		engine: e,
		cfg:    cfg,
		flags:  flags,
	}
}

func (p *Parser) SetEngine(e Engine) {
	p.engine = e
}

func (p *Parser) SetConfig(cfg *config.Config) {
	p.cfg = cfg
}

func (p *Parser) SetFlags(flags *config.Flags) {
	p.flags = flags
}

func (p *Parser) Parse() error {
	return p.engine.Parse(p.cfg, p.flags)
}
