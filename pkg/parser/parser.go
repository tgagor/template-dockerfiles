package parser

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"td/pkg/cmd"
	"td/pkg/config"
	"td/pkg/runner"
)

func Run(workdir string, cfg *config.Config, flag config.Flags) error {
	for name, img := range cfg.Images {
		slog.Debug("Analyzing", "image", name, "config", img)
		dockerfileTemplate := filepath.Join(workdir, img.Dockerfile)
		slog.Debug("Loading dockerfile: " + dockerfileTemplate)
		if img.Excludes != nil {
			slog.Debug("Excluded config sets", "excludes", img.Excludes)
		}

		interfaceVariables := convertToInterfaceMap(img.Variables)
		combinations := getCombinations(interfaceVariables)

		// var labels []string
		var tempFiles []string

		buildTasks := runner.New().Threads(flag.Threads).DryRun(flag.DryRun)
		// labelling have to happen in order, so no parallelism
		labelingTasks := runner.New().DryRun(flag.DryRun)
		pushTasks := runner.New().Threads(flag.Threads).DryRun(flag.DryRun)
		cleanupTasks := runner.New().Threads(flag.Threads).DryRun(flag.DryRun)
		for _, configSet := range combinations {
			// FIXME: This way of setting variables might collide with overrides
			// 		  set in "variables" section, I need to change order here.
			//		  New Map should be created with "config defaults", then
			//		  current configSet applied over it, and merged with cfg.
			configSet["tag"] = flag.Tag
			configSet["registry"] = cfg.Registry
			configSet["prefix"] = cfg.Prefix
			configSet["maintainer"] = cfg.Maintainer
			slog.Info("Building", "image", name, "config set", configSet)

			if isExcluded(configSet, img.Excludes) {
				slog.Debug("Skipping excluded", "config set", configSet, "excludes", img.Excludes)
				continue // break here, this set is excluded
			}

			// Collect all required data
			labels, err := templateLabels(img.Labels, configSet)
			if err != nil {
				return err
			}
			slog.Debug("Generated labels: " + strings.Join(labels, ", "))

			dockerfile := getDockerfilePath(dockerfileTemplate, configSet)
			slog.Debug("Generating temporary Dockerfile: " + dockerfile)
			tempFiles = append(tempFiles, dockerfile)
			slog.Debug("Tempfiles", "files", tempFiles)
			// name required to avoid collisions between images
			currentImage := name + "-" + getCombinationString(configSet)
			if !flag.DryRun {
				if err := templateFile(dockerfileTemplate, dockerfile, configSet); err != nil {
					return err
				}
			}

			// collect building image commands
			builder := cmd.New("docker").
				Arg("build").
				Arg("-f", dockerfile).
				Arg("-t", currentImage).
				Arg(getOCILabels(configSet)...).
				// TODO: Add open container labels automatically
				Arg(filepath.Dir(dockerfileTemplate)).
				SetVerbose(flag.Verbose)
			buildTasks = buildTasks.AddTask(builder)

			// collect labelling commands to keep order
			for _, l := range labels {
				labeler := cmd.New("docker").
					Arg("tag").
					Arg(currentImage).
					Arg(imageName(cfg.Registry, cfg.Prefix, l)).
					SetVerbose(flag.Verbose)
				labelingTasks = labelingTasks.AddTask(labeler)

				pusher := cmd.New("docker").
					Arg("push").
					Arg(imageName(cfg.Registry, cfg.Prefix, l))
				if !flag.Verbose { // TODO: check it
					pusher.Arg("--quiet")
				}
				pushTasks = pushTasks.AddTask(pusher)
			}

			// remove temporary labels
			dropTempLabel := cmd.New("docker").
				Arg("image", "rm", "-f").
				Arg(currentImage).
				SetVerbose(flag.Verbose)
			cleanupTasks = cleanupTasks.AddTask(dropTempLabel)
		}

		buildTasks.RunParallel()
		labelingTasks.Run()
		cleanupTasks.RunParallel()
		if flag.Push {
			slog.Info("Pushing images...")
			pushTasks.RunParallel()
		}

		// Cleanup temporary files
		for _, file := range tempFiles {
			if !flag.DryRun {
				defer removeFile(file)
			}
		}

		fmt.Println("")

	}
	return nil
}

// getCombinations generates all combinations of variables
func getCombinations(variables map[string][]interface{}) []map[string]interface{} {
	// Extract keys
	keys := make([]string, 0, len(variables))
	values := make([][]interface{}, 0, len(variables))

	// Collect keys and corresponding value slices
	for key, val := range variables {
		keys = append(keys, key)
		values = append(values, val)
	}

	// Resulting combinations
	var combinations []map[string]interface{}

	// Recursive helper to generate combinations
	var generate func(int, map[string]interface{})
	generate = func(depth int, current map[string]interface{}) {
		if depth == len(keys) {
			// Create a copy of the map and append it to the results
			combination := make(map[string]interface{}, len(current))
			for k, v := range current {
				combination[k] = v
			}
			combinations = append(combinations, combination)
			return
		}

		// Iterate over values for the current key
		key := keys[depth]
		for _, value := range values[depth] {
			current[key] = value
			generate(depth+1, current)
		}
	}

	// Start generating combinations
	generate(0, make(map[string]interface{}))

	return combinations
}

// this allows to treat both strings and integers the same
func convertToInterfaceMap(input map[string][]string) map[string][]interface{} {
	output := make(map[string][]interface{}, len(input))
	for key, values := range input {
		// Convert each []string to []interface{}
		interfaces := make([]interface{}, len(values))
		for i, v := range values {
			interfaces[i] = v
		}
		output[key] = interfaces
	}
	return output
}

func templateLabels(labelTemplates []string, configSet map[string]interface{}) ([]string, error) {
	var labels []string

	for _, label := range labelTemplates {
		templated, err := templateString(label, configSet)
		if err != nil {
			return nil, err
		}
		labels = append(labels, templated)
	}

	return labels, nil
}

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
		slog.Error("Failed to create a file: "+templateFile, "error", err)
		return err
	}
	defer f.Close()

	// var w io.Writer = f
	// if isDebugLevel() {
	// 	slog.Debug("HUGE DEBUG")
	// 	w = io.MultiWriter(os.Stdout, f)
	// }

	// Render templates using variables
	if err := t.Execute(f, args); err != nil {
		slog.Error("Failed to template file: "+templateFile, "error", err)
		return err
	}

	return nil
}

func sanitizeForFileName(input string) string {
	// Replace any character that is not a letter, number, or safe symbol (-, _) with an underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_]+`)
	return reg.ReplaceAllString(input, "_")
}

func getCombinationString(configSet map[string]interface{}) string {
	var parts []string
	for k, v := range configSet {
		if !ignoredKey(k) {
			// Apply sanitization to both key and value
			safeKey := sanitizeForFileName(k)
			safeValue := sanitizeForFileName(fmt.Sprintf("%v", v))
			parts = append(parts, fmt.Sprintf("%s-%s", safeKey, safeValue))
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, "-")
}

func getDockerfilePath(dockerFileTemplate string, configSet map[string]interface{}) string {
	dirname := filepath.Dir(dockerFileTemplate)
	filename := getCombinationString(configSet) + ".Dockerfile"
	return filepath.Join(dirname, filename)
}

func imageName(registry string, prefix string, name string) string {
	return path.Join(registry, prefix, name)
}

func ignoredKey(key string) bool {
	switch key {
	case
		"registry",
		"prefix",
		"maintainer",
		"tag":
		return true
	}
	return false
}

func CopyMapNoTag(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		if ignoredKey(k) {
			continue
		}
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = CopyMapNoTag(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}

func excludesToInterfaceMap(input []map[string]string) []map[string]interface{} {
	output := make([]map[string]interface{}, len(input))
	for _, o := range input {
		// Convert each []string to []interface{}
		interfaces := make(map[string]interface{}, len(o))
		for k, v := range o {
			interfaces[k] = v
		}
		output = append(output, interfaces)
	}
	return output
}

func isExcluded(item map[string]interface{}, excludes []map[string]string) bool {
	copy := CopyMapNoTag(item)
	for _, e := range excludesToInterfaceMap(excludes) {
		if reflect.DeepEqual(copy, e) {
			return true
		}
	}
	return false
}

func removeFile(file string) {
	slog.Debug("Removing temporary file: " + file)
	if err := os.Remove(file); err != nil {
		slog.Error("Failed to remove file", slog.Any("error", err))
	}
}
