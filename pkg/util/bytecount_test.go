package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

func TestByteCountIEC(t *testing.T) {
	// Arrange
	input := []uint64{
		1000,
		1024,
		1000 * 1000,
		1024 * 1024,
		1000 * 1000 * 1000,
		1024 * 1024 * 1024,
	}
	expected := []string{
		"1000 B",
		"1.0 KiB",
		"976.6 KiB",
		"1.0 MiB",
		"953.7 MiB",
		"1.0 GiB",
	}

	// Assert
	for i, input := range input {
		assert.Equal(t, expected[i], util.ByteCountIEC(input))
	}
}

func TestByteCountSI(t *testing.T) {
	// Arrange
	input := []uint64{
		1000,
		1024,
		1000 * 1000,
		1024 * 1024,
		1000 * 1000 * 1000,
		1024 * 1024 * 1024,
	}
	expected := []string{
		"1.0 kB",
		"1.0 kB",
		"1.0 MB",
		"1.0 MB",
		"1.0 GB",
		"1.1 GB",
	}

	// Assert
	for i, input := range input {
		assert.Equal(t, expected[i], util.ByteCountSI(input))
	}
}
