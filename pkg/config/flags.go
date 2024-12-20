package config

type Flags struct {
	Build        bool
	BuildFile    string
	Delete       bool
	DryRun       bool
	PrintVersion bool
	Push         bool
	Squash       bool
	Tag          string
	Threads      int
	Verbose      bool
	Image        string
	Engine       string
}
