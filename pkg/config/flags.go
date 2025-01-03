package config

type Flags struct {
	Build        bool
	BuildFile    string
	Delete       bool
	DryRun       bool
	Engine       string
	Image        string
	NoColor      bool
	PrintVersion bool
	Push         bool
	Squash       bool
	Tag          string
	Threads      int
	Verbose      bool
	Debug        bool
}
