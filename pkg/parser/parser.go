package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog/log"

	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/runner"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

// TODO: add multi-arch building support
func Run(workdir string, cfg *config.Config, flag config.Flags) error {
	for _, name := range cfg.ImageOrder {
		// Limit building to a single image
		if flag.Image != "" && name != flag.Image {
			continue
		}

		img := cfg.Images[name]
		log.Debug().Str("image", name).Interface("config", img).Msg("Parsing")
		dockerfileTemplate := filepath.Join(workdir, img.Dockerfile)
		log.Debug().Str("dockerfile", dockerfileTemplate).Msg("Processing")
		if img.Excludes != nil {
			log.Debug().Interface("excludes", img.Excludes).Msg("Excluded config sets")
		}

		var toSquash []string

		buildTasks := runner.New().Threads(flag.Threads).DryRun(!flag.Build)
		// labelling have to happen in order, so no parallelism
		taggingTasks := runner.New().DryRun(!flag.Build)
		pushTasks := runner.New().Threads(flag.Threads).DryRun(!flag.Build)
		cleanupTasks := runner.New().Threads(flag.Threads).DryRun(!flag.Build)

		combinations := getCombinations(img.Variables)
		for _, configSet := range combinations {
			log.Info().Str("image", name).Msg("Building")
			// FIXME: This way of setting variables might collide with overrides
			// 		  set in "variables" section, I need to change order here.
			//		  New Map should be created with "config defaults", then
			//		  current configSet applied over it, and merged with cfg.
			configSet["image"] = name
			configSet["tag"] = flag.Tag
			configSet["registry"] = cfg.Registry
			configSet["prefix"] = cfg.Prefix
			configSet["maintainer"] = cfg.Maintainer
			configSet["labels"] = make(map[string]string)
			maps.Copy(configSet["labels"].(map[string]string), cfg.GlobalLabels)
			maps.Copy(configSet["labels"].(map[string]string), img.Labels)

			// skip excluded config sets
			if isExcluded(configSet, img.Excludes) {
				log.Debug().Interface("config set", configSet).Interface("excludes", img.Excludes).Msg("Skipping excluded")
				continue
			}

			// Collect all required data
			tags := collectTags(img, configSet, name)

			// Collect labels, starting with global labels, then oci, then per image
			labels := getOCILabels(configSet)
			maps.Copy(labels, collectLabels(configSet))

			var dockerfile string
			if strings.HasSuffix(dockerfileTemplate, ".tpl") {
				dockerfile = getDockerfilePath(dockerfileTemplate, name, configSet)
				log.Debug().Str("dockerfile", dockerfile).Msg("Generating temporary")

				// Template Dockerfile
				err := templateFile(dockerfileTemplate, dockerfile, configSet)
				util.FailOnError(err)

				// Cleanup temporary files
				if flag.Delete {
					defer removeFile(dockerfile)
				}
			} else {
				dockerfile = dockerfileTemplate
			}

			// name is required to avoid collisions between images or
			// when variables are not defined to have actual image name
			currentImage := strings.Join([]string{name, getCombinationString(configSet)}, "-")

			// collect building image commands
			builder := cmd.New("docker").Arg("build").
				Arg("-f", dockerfile).
				Arg("-t", currentImage).
				Arg(labelsToArgs(labels)...).
				Arg(filepath.Dir(dockerfileTemplate)).
				SetVerbose(flag.Verbose)
			buildTasks = buildTasks.AddTask(builder)

			// squash if demanded
			if flag.Squash {
				toSquash = append(toSquash, currentImage)
			}

			// collect tagging commands to keep order
			for _, t := range tags {
				taggedImg := imageName(cfg.Registry, cfg.Prefix, t)
				tagger := cmd.New("docker").Arg("tag").
					Arg(currentImage).
					Arg(taggedImg).
					SetVerbose(flag.Verbose).
					PreInfo("Tagging " + taggedImg)
				taggingTasks = taggingTasks.AddTask(tagger)

				pusher := cmd.New("docker").Arg("push").
					Arg(taggedImg).
					PreInfo("Pushing " + taggedImg)
				if !flag.Verbose { // TODO: check it
					pusher = pusher.Arg("--quiet")
				}
				pushTasks = pushTasks.AddTask(pusher)
			}

			// remove temporary labels
			dropTempLabel := cmd.New("docker").Arg("image", "rm", "-f").
				Arg(currentImage).
				SetVerbose(flag.Verbose)
			cleanupTasks = cleanupTasks.AddTask(dropTempLabel)
		}

		if flag.Build {
			err := buildTasks.Run()
			util.FailOnError(err, "Building failed with error, check error above. Exiting.")
		}

		// let squash it
		if flag.Build && flag.Squash {
			squashImages(flag, toSquash)
		}

		// continue classical build
		if flag.Build {
			err := taggingTasks.Run()
			util.FailOnError(err, "Tagging failed with error, check error above. Exiting.")
			err = cleanupTasks.Run()
			util.FailOnError(err, "Dropping temporary images failed. Exiting.")
		}
		if flag.Push {
			err := pushTasks.Run()
			util.FailOnError(err, "Pushing images failed, check error above. Exiting.")
		}

		fmt.Println("")

	}
	return nil
}

func squashImages(flag config.Flags, toSquash []string) {
	runImages := runner.New().Threads(flag.Threads).DryRun(!flag.Build)
	exportImages := runner.New().Threads(flag.Threads).DryRun(!flag.Build)
	removeDeadContainers := runner.New().Threads(flag.Threads).DryRun(!flag.Build)
	importTarsToImgs := runner.New().Threads(flag.Threads).DryRun(!flag.Build)

	var squashed []string

	for _, img := range toSquash {
		sanitizedImg := sanitizeForFileName(img)

		runItFirst := cmd.New("docker").
			Arg("run").
			Arg("--name", sanitizedImg).
			Arg(img).
			Arg("true").
			SetVerbose(flag.Verbose)
		runImages = runImages.AddTask(runItFirst)

		imgMetadata, err := inspectImg(img)
		util.FailOnError(err, "Couldn't inspect Docker image.")
		log.Debug().Interface("data", imgMetadata).Msg("Docker inspect result")

		tmpTarFile := sanitizedImg + ".tar"
		exportIt := cmd.New("docker").
			Arg("export").
			Arg(sanitizedImg).
			Arg("-o", tmpTarFile).
			PreInfo(fmt.Sprintf("Squashing %s of size: %s", img, ByteCountIEC(imgMetadata[0].Size))).
			SetVerbose(flag.Verbose)
		exportImages = exportImages.AddTask(exportIt)
		dropIt := cmd.New("docker").Arg("rm").Arg("-f").Arg(sanitizedImg)
		removeDeadContainers = removeDeadContainers.AddTask(dropIt)

		importIt := cmd.New("docker").Arg("import")
		for _, item := range imgMetadata {
			// paring ENV
			for _, env := range item.Config.Env {
				importIt = importIt.Arg("--change", "ENV "+env)
			}

			// parsing CMD
			if command, err := json.Marshal(item.Config.Cmd); err != nil {
				log.Error().Err(err).Str("image", img).Msg("Can't parse CMD")
			} else {
				importIt = importIt.Arg("--change", "CMD "+string(command))
			}

			// parsing VOLUME
			if vol, err := json.Marshal(item.Config.Volumes); err != nil {
				log.Error().Err(err).Str("image", img).Msg("Can't parse VOLUME")
			} else {
				importIt = importIt.Arg("--change", "VOLUME "+string(vol))
			}

			// parsing LABELS
			for key, value := range item.Config.Labels {
				importIt = importIt.Arg("--change", fmt.Sprintf("LABEL %s=\"%s\"", key, strings.ReplaceAll(value, "\n", "")))
			}

			// parsing ENTRYPOINT
			if entrypoint, err := json.Marshal(item.Config.Entrypoint); err != nil {
				log.Error().Err(err).Str("image", img).Msg("Can't parse ENTRYPOINT")
			} else {
				importIt = importIt.Arg("--change", "CMD "+string(entrypoint))
			}

			// parsing WORKDIR
			if item.Config.WorkingDir != "" {
				importIt = importIt.Arg("--change", "WORKDIR "+item.Config.WorkingDir)
			}
		}
		importIt = importIt.Arg(tmpTarFile).Arg(sanitizedImg).
			SetVerbose(flag.Verbose)
		importTarsToImgs = importTarsToImgs.AddTask(importIt)

		squashed = append(squashed, sanitizedImg)
		defer removeFile(tmpTarFile)
	}

	err := runImages.Run()
	util.FailOnError(err)
	err = exportImages.Run()
	util.FailOnError(err)
	err = removeDeadContainers.Run()
	util.FailOnError(err)
	err = importTarsToImgs.Run()
	util.FailOnError(err)

	for _, img := range squashed {
		imgMetadata, err := inspectImg(img)
		util.FailOnError(err, "Couldn't inspect Docker image.")
		log.Debug().Interface("data", imgMetadata).Msg("Docker inspect result")
		log.Info().Msg(fmt.Sprintf("Squashed %s to size: %s", img, ByteCountIEC(imgMetadata[0].Size)))
	}
}

func collectLabels(configSet map[string]interface{}) map[string]string {
	labels, err := templateLabels(configSet["labels"].(map[string]string), configSet)
	util.FailOnError(err)
	if len(labels) > 0 {
		log.Info().Interface("labels", labels).Msg("Generating")
	}
	return labels
}

func collectTags(img config.ImageConfig, configSet map[string]interface{}, name string) []string {
	tags, err := templateTags(img.Tags, configSet)
	util.FailOnError(err)
	if len(tags) > 0 {
		log.Info().Interface("tags", tags).Msg("Generating")
	} else {
		log.Error().Str("image", name).Msg("No 'tags' defined for")
		log.Error().Msg("Building without 'tags', would just overwrite images in place, which is pointless. Add 'tags' block to continue.")
		os.Exit(1)
	}
	return tags
}

// generates all combinations of variables
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

func templateTags(tagTemplates []string, configSet map[string]interface{}) ([]string, error) {
	var tags []string

	for _, label := range tagTemplates {
		templated, err := templateString(label, configSet)
		if err != nil {
			return nil, err
		}
		tags = append(tags, templated)
	}

	return tags, nil
}

func templateLabels(labelTemplates map[string]string, configSet map[string]interface{}) (map[string]string, error) {
	labels := map[string]string{}

	for label, value := range labelTemplates {
		templatedLabel, err := templateString(label, configSet)
		if err != nil {
			return nil, err
		}
		templatedValue, err := templateString(value, configSet)
		if err != nil {
			return nil, err
		}
		labels[templatedLabel] = templatedValue
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
		log.Error().Err(err).Str("file", templateFile).Msg("Failed to create")
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
		log.Error().Err(err).Str("file", templateFile).Msg("Failed to template")
		return err
	}

	return nil
}

func sanitizeForFileName(input string) string {
	// Replace any character that is not a letter, number, or safe symbol (-, _) with an underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_\.]+`)
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

func getDockerfilePath(dockerFileTemplate string, image string, configSet map[string]interface{}) string {
	dirname := filepath.Dir(dockerFileTemplate)
	filename := strings.Join([]string{image, getCombinationString(configSet) + ".Dockerfile"}, "-")
	return filepath.Join(dirname, sanitizeForFileName(filename))
}

func imageName(registry string, prefix string, name string) string {
	return path.Join(registry, prefix, name)
}

func ignoredKey(key string) bool {
	switch key {
	case
		"image",
		"registry",
		"prefix",
		"maintainer",
		"tag",
		"labels":
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
	log.Debug().Str("file", file).Msg("Removing temporary")
	if err := os.Remove(file); err != nil {
		log.Error().Err(err).Str("file", file).Msg("Failed to remove")
	}
}
