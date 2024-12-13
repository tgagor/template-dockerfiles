package config

type Flags struct {
	BuildFile    string
	DryRun       bool
	Push         bool
	Threads      int
	Tag          string
	Verbose      bool
	PrintVersion bool
}
