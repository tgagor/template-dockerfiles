package cmd

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

type Cmd struct {
	cmd      string
	args     []string
	verbose  bool
	preText  string
	postText string
}

func New(c string) Cmd {
	return Cmd{
		cmd:      c,
		verbose:  false,
		preText:  "",
		postText: "",
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

func (c Cmd) PreInfo(msg string) Cmd {
	c.preText = msg
	return c
}

func (c Cmd) PostInfo(msg string) Cmd {
	c.postText = msg
	return c
}

func (c Cmd) Run(ctx context.Context) error {
	if c.preText != "" {
		log.Info().Msg(c.preText)
	}

	cmd := exec.CommandContext(ctx, c.cmd, c.args...)

	// pipe the commands output to the applications
	if c.verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	log.Debug().Str("cmd", c.cmd).Interface("args", c.args).Msg("Running")
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Msg("Could not run command")
		return err
	}

	if c.postText != "" {
		log.Info().Msg(c.postText)
	}
	return nil
}

func (c Cmd) String() string {
	return c.cmd + " " + strings.Join(c.args, " ")
}
