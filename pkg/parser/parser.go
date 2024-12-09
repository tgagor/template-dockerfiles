package parser

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"template-dockerfiles/pkg/cmd"
	"template-dockerfiles/pkg/config"
	"template-dockerfiles/pkg/runner"
)

func Run(workdir string, cfg *config.Config, args map[string]any) error {
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
		buildTasks := runner.New().Threads(args["threads"]).DryRun(args.dryRun)
		labelingTasks := runner.New().Threads(args["threads"]).DryRun(args.dryRun)
		cleanupTasks := runner.New().Threads(args["threads"].(int)).DryRun(args.dryRun)
		for _, configSet := range combinations {
			configSet["tag"] = tag
			slog.Debug("Per image", "config set", configSet)
			slog.Info("Building", "image", name)

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
			tempFiles = append(tempFiles, dockerfile)
			slog.Debug("Generating temporary Dockerfile: " + dockerfile)

			// registry := cfg.Registry
			// prefix = cfg.Prefix
			// image_params = collect_params(config_set, playbook)
			// templated_dockerfile = template_file(template_path, image_params)
			// labels = collect_labels(config_set, params["labels"])
			// dockerfile = get_dockerfile_path(template_path, config_set)
			// temp_files.append(dockerfile)  # for later cleanup

			currentImage := getCombination(configSet)

			// collect building image commands
			builder := cmd.New("docker").
				Arg("build").
				Arg("-f", dockerfile).
				Arg("-t", currentImage).
				// TODO: Add opencontainer labels automatically
				Arg(filepath.Dir(dockerfileTemplate))
			buildTasks = buildTasks.AddTask(builder)

			// collect labelling commands to keep order
			for _, l := range labels {

				labeler := cmd.New("docker").
					Arg("tag").
					Arg(currentImage).
					Arg(imageName(cfg.Registry, cfg.Prefix, l))
				labelingTasks = labelingTasks.AddTask(labeler)
			}

			// collect cleanup tasks
			dropTempLabel := cmd.New("docker").
				Arg("image", "rm").
				Arg(currentImage)
			cleanupTasks = cleanupTasks.AddTask(dropTempLabel)

			if err := templateFile(dockerfileTemplate, dockerfile, configSet); err != nil {
				return err
			}
		}

		buildTasks.Run()
		labelingTasks.Run()
		cleanupTasks.Run()

		// Cleanup temporary files
		for _, file := range tempFiles {
			slog.Debug("Removing temporary file: " + file)
			if err := os.Remove(file); err != nil {
				slog.Error("Failed to remove file", slog.Any("error", err))
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
	t := template.Must(template.New(pattern).Parse(pattern))
	if err := t.Execute(&output, args); err != nil {
		return "", err
	}

	return output.String(), nil
}

func templateFile(templateFile string, destinationFile string, args map[string]interface{}) error {
	t, err := template.ParseFiles(templateFile)
	if err != nil {
		slog.Error("Failed to parse file: "+templateFile, "error", err)
		return err
	}

	f, err := os.Create(destinationFile)
	if err != nil {
		slog.Error("Failed to create a file: "+templateFile, "error", err)
		return err
	}

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

	return f.Close()
}

func getCombination(configSet map[string]interface{}) string {
	var parts []string
	for k, v := range configSet {
		parts = append(parts, fmt.Sprintf("%s-%s", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, "-")
}

func getDockerfilePath(dockerFileTemplate string, configSet map[string]interface{}) string {
	dirname := filepath.Dir(dockerFileTemplate)
	filename := getCombination(configSet) + ".Dockerfile"
	return filepath.Join(dirname, filename)
}

func imageName(registry string, prefix string, name string) string {
	return filepath.Join(registry, prefix, name)
}

func CopyMapNoTag(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		if k == "tag" {
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
	slog.Debug("Result: false")
	return false
}
