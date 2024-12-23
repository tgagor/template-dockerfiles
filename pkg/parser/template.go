package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog/log"
)

func templateString(pattern string, args map[string]interface{}) (string, error) {
	var output bytes.Buffer
	t := template.Must(template.New(pattern).Funcs(sprig.TxtFuncMap()).Parse(pattern))
	if err := t.Execute(&output, args); err != nil {
		return "", err
	}

	return output.String(), nil
}

func templateFile(templateFile string, destinationFile string, args map[string]interface{}) error {
	t := template.Must(
		template.New(filepath.Base(templateFile)).Funcs(sprig.TxtFuncMap()).ParseFiles(templateFile),
	)

	f, err := os.Create(destinationFile)
	if err != nil {
		log.Error().Err(err).Str("file", templateFile).Msg("Failed to create")
		return err
	}
	defer f.Close()

	// Render templates using variables
	if err := t.Execute(f, args); err != nil {
		log.Error().Err(err).Str("file", templateFile).Msg("Failed to template")
		return err
	}

	return nil
}
