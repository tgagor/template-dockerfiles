package tests_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
)

func cmd(args ...string) shell.Command {
	defaultArgs := []string{}
	return shell.Command{
		Command: "../bin/td",
		Args:    append(defaultArgs, args...),
		Logger:  logger.Discard,
	}
}

// Simplest possible test, just print version and exit
// Should print version to stdout
// Should not fail
func TestPrintVersion(t *testing.T) {
	t.Parallel()

	cmd := cmd("-V")

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.NotNil(t, out) // should print version
	assert.Nil(t, err)
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}

func TestFailWithoutConfigParam(t *testing.T) {
	t.Parallel()

	// missing --config
	cmd := cmd("--build")

	// should fail with error
	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.NotNil(t, err)
	assert.Contains(t, out, "the --config flag is required")

	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.NotEqual(t, code, 0)
}

// Simple build o 3 images
// Variable with space should be properly escaped
//
//	and do not fail the build
//
// 3 files should be created
func TestCase1(t *testing.T) {
	t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-1.yaml",
		"--tag", "v1.1.1",
	)

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Regexp(t, "Building.*image=test-case-1", out)

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

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Regexp(t, "Building.*image=test-case-2", out)
	assert.Regexp(t, "Building.*image=test-case-2b", out)

	// not all of them match but should
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

func TestCase3(t *testing.T) {
	t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-3.yaml",
		"--tag", "v3.3.3",
	)

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.NotNil(t, err) // should fail
	assert.Regexp(t, "Building.*image=test-case-3", out)

	assert.Contains(t, out, "No 'tags' defined for image=test-case-3")
	assert.Contains(t, out, "building without 'tags', would just overwrite images in place, which is pointless - add 'tags' block to continue")

	// should fail
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.NotEqual(t, code, 0)
}

func TestCase4(t *testing.T) {
	t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-4.yaml",
		"--tag", "v4.4.4",
		"--delete",
	)

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Regexp(t, "Building.*image=test-case-4", out)

	// do not fail
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}

// This test verifies if trimming of tags and labels works
func TestCase5(t *testing.T) {
	t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-5.yaml",
		"--tag", "v5.5.5",
		"--delete",
	)

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Regexp(t, "Building.*image=test-case-5", out)

	// do not fail
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}

func TestCase6(t *testing.T) {
	t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-6.yaml",
		"--tag", "v6.6.6",
		"--delete",
	)

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Regexp(t, "Building.*image=test-case-6", out)

	// do not fail
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}

func TestCase7(t *testing.T) {
	t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-7.yaml",
		"--delete",
	)

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Regexp(t, "Building.*image=test-case-7", out)

	// do not fail
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}

func TestCase8(t *testing.T) {
	t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-8.yaml",
		"--tag", "v8.8.8",
		"--delete",
	)

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Regexp(t, "Building.*image=test-case-8", out)
	assert.Contains(t, out, "Skipping excluded config set=")

	// do not fail
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}

func TestCase9(t *testing.T) {
	t.Parallel()

	cmd := cmd(
		"--no-color",
		"--config", "test-9.yaml",
		"--tag", "v9.9.9",
	)

	out, err := shell.RunCommandAndGetOutputE(t, cmd)
	assert.Nil(t, err)
	assert.Regexp(t, "Building.*image=test-case-9", out)

	// should generage args
	// FIXME: amend after changes
	// assert.Contains(t, out, "Generating args={\"BASEIMAGE\":\"3.20\",\"TIMEZONE\":\"UTC\"}")
	// assert.Contains(t, out, "Generating args={\"BASEIMAGE\":\"3.20\",\"TIMEZONE\":\"EST\"}")
	// assert.Contains(t, out, "Generating args={\"BASEIMAGE\":\"3.21\",\"TIMEZONE\":\"UTC\"}")
	// assert.Contains(t, out, "Generating args={\"BASEIMAGE\":\"3.21\",\"TIMEZONE\":\"EST\"}")

	// should not create temporary Dockerfiles
	assert.False(t, files.FileExists("test-case-9-alpine-3.20-timezone-EST.Dockerfile"))
	assert.False(t, files.FileExists("test-case-9-alpine-3.20-timezone-UTC.Dockerfile"))
	assert.False(t, files.FileExists("test-case-9-alpine-3.21-timezone-EST.Dockerfile"))
	assert.False(t, files.FileExists("test-case-9-alpine-3.21-timezone-UTC.Dockerfile"))

	// do not fail
	code, err := shell.GetExitCodeForRunCommandError(err)
	assert.Nil(t, err)
	assert.Equal(t, code, 0)
}
