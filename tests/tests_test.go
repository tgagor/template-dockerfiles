package tests_test

import (
	// "bytes"
	// "fmt"
	// "regexp"
	// "strings"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// "github.com/stretchr/testify/require"

	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
)

// var cmd shell.Command

// func TestMain(m *testing.M) {

//     exitVal := m.Run()

//     os.Exit(exitVal)
// }

func cmd(args ...string) shell.Command {
	defaultArgs := []string{}
	return shell.Command{
		Command: "../bin/td",
		Args:    append(defaultArgs, args...),
		Logger:  logger.Discard,
	}
}

func TestRunTDVersion(t *testing.T) {
	t.Parallel()

	cmd := cmd("-V")

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Contains(t, out, "development")
	assert.Nil(t, err)
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}

func TestCase1GenerateFiles(t *testing.T) {
	// t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-1.yaml",
		"--tag", "v1.1.1",
	)

	// returns what I want
	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Contains(t, out, "Building image=test-case-1")
	assert.Contains(t, out, "Generating tags=[\"test-case-1:v1.1.1-alpine3.18\",\"test-case-1:alpine3\",\"test-case-1\"]")
	assert.Contains(t, out, "Generating tags=[\"test-case-1:v1.1.1-alpine3.19\",\"test-case-1:alpine3\",\"test-case-1\"]")
	assert.Contains(t, out, "Generating tags=[\"test-case-1:v1.1.1-alpine3.20\",\"test-case-1:alpine3\",\"test-case-1\"]")

	// command should not fail
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)

	// should create 3 files
	dockerFiles := []string{
		"test-case-1-alpine-3.18-multiword-kuku-ruku.Dockerfile",
		"test-case-1-alpine-3.19-multiword-kuku-ruku.Dockerfile",
		"test-case-1-alpine-3.20-multiword-kuku-ruku.Dockerfile",
	}
	for _, f := range dockerFiles {
		assert.True(t, files.FileExists(f))

		// cleanup automatically
		require.NoError(t, os.Remove(f))
	}

}

func TestCase2(t *testing.T) {
	t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-2.yaml",
		"--tag", "v2.2.2",
		"--delete",
		"--verbose",
	)

	// returns what I want
	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Contains(t, out, "Building image=test-case-2")
	assert.Contains(t, out, "Building image=test-case-2b")

	assert.Contains(t, out, "\"maintainer\":\"Tomasz Gągor <tomasz@gagor.pl>\"")
	assert.Contains(t, out, "\"org.opencontainers.image.authors\":\"Tomasz Gągor <tomasz@gagor.pl>\"")
	// assert.Contains(t, out, "org.opencontainers.image.branch")
	assert.Contains(t, out, "org.opencontainers.image.created")
	assert.Contains(t, out, "org.opencontainers.image.description")
	assert.Contains(t, out, "\"org.opencontainers.image.licenses\":\"GPL-2.0-only\"")
	// assert.Regexp(t, "\"org.opencontainers.image.revision\":\".*\"", out)
	// assert.Contains(t, out, "\"org.opencontainers.image.source\":\"git@github.com:tgagor/docker-templater.git\"")
	assert.Contains(t, out, "\"org.opencontainers.image.vendor\":\"Test Corp\"")
	assert.Contains(t, out, "\"org.opencontainers.image.version\":\"v2.2.2\"")

	// validate file deletion to work
	assert.Contains(t, out, "Templated Dockerfiles will be deleted at end")
	assert.Contains(t, out, "Removing temporary file=test-case-2b-alpine-3.20.Dockerfile")
	assert.Contains(t, out, "Removing temporary file=test-case-2-alpine-3.20.Dockerfile")
	assert.Contains(t, out, "Removing temporary file=test-case-2-alpine-3.19.Dockerfile")
	assert.Contains(t, out, "Removing temporary file=test-case-2-alpine-3.18.Dockerfile")

	// do not fail
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}
