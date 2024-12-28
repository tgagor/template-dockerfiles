package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tgagor/template-dockerfiles/pkg/cmd"
)

func TestRunner(t *testing.T) {
	// Arrange
	input := []string{
		cmd.New("echo").Arg("hello").Arg("world").String(),
		cmd.New("cmd-only").String(),
		cmd.New("").String(),
	}
	expected := []string{
		"echo hello world",
		"cmd-only",
		"",
	}

	// Assert
	for i, input := range input {
		assert.Equal(t, expected[i], input)
	}
}
