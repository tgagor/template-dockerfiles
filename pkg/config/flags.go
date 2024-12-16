package config

type Flags struct {
	BuildFile    string
	DryRun       bool
	Build        bool
	Push         bool
	Delete       bool
	Threads      int
	Tag          string
	Verbose      bool
	PrintVersion bool
}
