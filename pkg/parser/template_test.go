package parser_test

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/tgagor/template-dockerfiles/pkg/parser"
)

// disable logging as it's trashing test's output
func TestMain(m *testing.M) {
	// log.SetOutput(io.Discard)
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	m.Run()
}

func TestTemplateString(t *testing.T) {
	t.Parallel()

	// Arrange
	inputStrings := []string{
		"{{ .key }}",
		"{{ .key }}",
		"{{ .key }}",
		"  {{ .key }}  ",
		"{{ .sprig | default \"works\" }}",
		"{{range .loop}}{{.}}{{ end }}",
	}
	inputArgs := []map[string]interface{}{
		{"key": "value"},
		{"key": 1},
		{"key": 1.43},
		{"key": "value"},
		{"sprig": ""},
		{"loop": []int{1, 2, 3}},
	}

	expected := []string{
		"value",
		"1",
		"1.43",
		"  value  ",
		"works",
		"123",
	}

	// Assert
	for i, input := range inputStrings {
		result, _ := parser.TemplateString(input, inputArgs[i])
		assert.Equal(t, expected[i], result)
	}
}

func TestTemplateList(t *testing.T) {
	t.Parallel()

	input := []string{
		"test-case-1:{{ .tag }}-alpine{{ .alpine }}",
		"test-case-1:alpine{{ .alpine | splitList \".\" | first }}",
		"   test-case-1   \n",
	}
	configSet := map[string]interface{}{
		"alpine": "3.33",
		"tag":    "version from param",
	}

	expected := []string{
		"test-case-1:version from param-alpine3.33",
		"test-case-1:alpine3",
		"test-case-1",
	}

	result, err := parser.TemplateList(input, configSet)
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
}

func TestTemplateMap(t *testing.T) {
	t.Parallel()

	input := map[string]string{
		"org.opencontainers.image.description":        "Alpine Linux {{ .alpine }}\nVersion {{ .tag }}",
		"\norg.opencontainers.image.nama   ":          "alpine:{{ .alpine }}",
		"org.opencontainers.image.{{ .alpine }}.nama": "   alpine:{{ .alpine }}  \n",
	}
	configSet := map[string]interface{}{
		"alpine": "3.33",
		"tag":    "version from param",
	}

	expected := map[string]string{
		"org.opencontainers.image.description": "Alpine Linux 3.33\nVersion version from param",
		"org.opencontainers.image.nama":        "alpine:3.33",
		"org.opencontainers.image.3.33.nama":   "alpine:3.33",
	}

	result, err := parser.TemplateMap(input, configSet)
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
}
