package parser

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/config"
)

type Image struct {
	Name       string
	Registry   string
	Prefix     string
	Dockerfile string
	Tags       []string
	Labels     map[string]string
	Args       map[string]string
	Platforms  []string
	Flags      *config.Flags
}

func init()

func New(flags *config.Flags) *Image {
	return &Image{
		Tags:      []string{},
		Labels:    map[string]string{},
		Args:      map[string]string{},
		Platforms: []string{},
		Flags:     flags,
	}
}

func From(configSet map[string]interface{}, flags *config.Flags) *Image {
	return &Image{
		Name:      configSet["name"].(string),
		Registry:  configSet["registry"].(string),
		Prefix:    configSet["prefix"].(string),
		Tags:      configSet["tags"].([]string),
		Labels:    configSet["labels"].(map[string]string),
		Args:      configSet["args"].(map[string]string),
		Platforms: configSet["platforms"].([]string),
		Flags:     flags,
	}
}

func (c *Image) SetName(name string) *Image {
	c.Name = name
	return c
}

func (c *Image) String() string {
	return c.FullName()
}

func (c *Image) FullName() string {
	return strings.ToLower(path.Join(c.Registry, c.Prefix, c.Name))
}

func (c *Image) SetRegistry(registry string) *Image {
	c.Registry = registry
	return c
}

func (c *Image) SetPrefix(prefix string) *Image {
	c.Prefix = prefix
	return c
}

func (c *Image) SetTags(tags ...[]string) *Image {
	c.Prefix = tags
	return c
}

func (c *Image) Equal(Image *Image) bool {
	return c.String() == Image.String()
}

func (c *Image) Arg(args ...string) *Image {
	c.args = append(c.args, args...)
	return c
}

func (c *Image) SetVerbose(verbosity bool) *Image {
	c.verbose = verbosity
	return c
}

func (c *Image) PreInfo(msg string) *Image {
	c.preText = msg
	return c
}

func (c *Image) PostInfo(msg string) *Image {
	c.postText = msg
	return c
}

func (c *Image) Run(ctx context.Context) (string, error) {
	if c.Image == "" {
		return "", errors.New("command not set")
	}
	if c.preText != "" {
		log.Info().Msg(c.preText)
	}

	Image := exec.CommandContext(ctx, c.Image, c.args...)

	// pipe the commands output to the applications
	var b bytes.Buffer
	if c.verbose {
		Image.Stdout = os.Stdout
		Image.Stderr = os.Stderr
	} else {
		Image.Stdout = &b
		Image.Stderr = &b
	}

	log.Debug().Str("Image", c.Image).Interface("args", c.args).Msg("Running")
	err := Image.Run()

	// Check for context cancellation or timeout
	if ctx.Err() != nil {
		// If the context was canceled, suppress output and return context error
		if ctx.Err() == context.Canceled {
			log.Warn().Str("Image", c.Image).Msg("Command was cancelled")
		} else if ctx.Err() == context.DeadlineExceeded {
			log.Warn().Str("Image", c.Image).Msg("Command timed out")
		}
		return "", ctx.Err()
	}

	// Handle other errors
	if err != nil {
		log.Error().Err(err).Str("Image", c.Image).Interface("args", c.args).Msg("Could not run command")
		// c.setOutput(&b)
		c.output = b.String()
		log.Error().Msg(c.output)
		return c.output, err
	}
	c.output = b.String()

	if c.postText != "" {
		log.Info().Msg(c.postText)
	}
	return c.output, nil
}

func (c *Image) Output() (string, error) {
	Image := exec.Command(c.Image, c.args...)

	// pipe the commands output to the applications
	var b bytes.Buffer
	if c.verbose {
		Image.Stdout = os.Stdout
		Image.Stderr = os.Stderr
	} else {
		Image.Stdout = &b
		Image.Stderr = &b
	}

	log.Debug().Str("Image", c.Image).Interface("args", c.args).Msg("Running")
	err := Image.Run()

	// Handle other errors
	if err != nil {
		log.Error().Err(err).Str("Image", c.Image).Interface("args", c.args).Msg("Could not run command")
		c.output = b.String()
		log.Error().Msg(c.output)
		return c.output, err
	}
	c.output = b.String()

	return c.output, nil
}
