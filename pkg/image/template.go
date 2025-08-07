package image

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog/log"
)

func TemplateString(pattern string, args map[string]interface{}) (string, error) {
	var output bytes.Buffer
	t := template.Must(template.New(pattern).Funcs(sprig.TxtFuncMap()).Parse(pattern))
	if err := t.Execute(&output, args); err != nil {
		return "", err
	}

	return output.String(), nil
}

func TemplateFile(templateFile string, destinationFile string, args map[string]interface{}) error {
	t := template.Must(
		template.New(filepath.Base(templateFile)).Funcs(sprig.TxtFuncMap()).ParseFiles(templateFile),
	)

	f, err := os.Create(destinationFile)
	if err != nil {
		log.Error().Err(err).Str("file", templateFile).Msg("Failed to create")
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing config file")
		}
	}()

	// Render templates using variables
	if err := t.Execute(f, args); err != nil {
		log.Error().Err(err).Str("file", templateFile).Msg("Failed to template")
		return err
	}

	return nil
}

func TemplateList(source []string, configSet map[string]interface{}) ([]string, error) {
	var templated []string

	for _, label := range source {
		templatedString, err := TemplateString(label, configSet)
		if err != nil {
			return nil, err
		}
		templated = append(templated, strings.Trim(templatedString, " \n"))
	}

	if len(templated) > 0 {
		log.Trace().Interface("source", source).Interface("templated", templated).Msg("Templating list")
	}

	return templated, nil
}

func TemplateMap(source map[string]string, configSet map[string]interface{}) (map[string]string, error) {
	templated := map[string]string{}

	for label, value := range source {
		templatedLabel, err := TemplateString(label, configSet)
		if err != nil {
			return nil, err
		}
		templatedValue, err := TemplateString(value, configSet)
		if err != nil {
			return nil, err
		}
		templatedLabel = strings.Trim(templatedLabel, " \n")
		templatedValue = strings.Trim(templatedValue, " \n")
		templated[templatedLabel] = templatedValue
	}

	if len(templated) > 0 {
		log.Trace().Interface("source", source).Interface("templated", templated).Msg("Templating map")
	}

	return templated, nil
}
