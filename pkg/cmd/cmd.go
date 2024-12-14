package cmd

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type Cmd struct {
	cmd     string
	args    []string
	verbose bool
}

func New(c string) Cmd {
	return Cmd{
		cmd: c,
	}
}

func (c Cmd) Equal(cmd Cmd) bool {
	return c.String() == cmd.String()
}

func (c Cmd) Arg(args ...string) Cmd {
	c.args = append(c.args, args...)
	return c
}

func (c Cmd) SetVerbose(verbosity bool) Cmd {
	c.verbose = verbosity
	return c
}

func (c Cmd) Run() error {
	cmd := exec.Command(c.cmd, c.args...)

	// pipe the commands output to the applications
	if c.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	slog.Debug("Running", "cmd", c.cmd, "args", c.args)
	if err := cmd.Run(); err != nil {
		slog.Error("Could not run command", "error", err)
		panic("Command " + cmd.String() + " failed!")
		// return err
	}
	return nil
}

func (c Cmd) String() string {
	return c.cmd + " " + strings.Join(c.args, " ")
}
