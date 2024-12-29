package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tgagor/template-dockerfiles/pkg/parser"
)

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
