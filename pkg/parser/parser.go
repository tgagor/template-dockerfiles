package parser

import (
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/tgagor/template-dockerfiles/pkg/builder"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

// TODO: add multi-arch building support
func Run(workdir string, cfg *config.Config, flags config.Flags) error {
	for _, name := range cfg.ImageOrder {
		// Limit building to a single image
		if flags.Image != "" && name != flags.Image {
			continue
		}

		img := cfg.Images[name]
		log.Debug().Str("image", name).Interface("config", img).Msg("Parsing")
		dockerfileTemplate := filepath.Join(workdir, img.Dockerfile)
		log.Debug().Str("dockerfile", dockerfileTemplate).Msg("Processing")
		if img.Excludes != nil {
			log.Debug().Interface("excludes", img.Excludes).Msg("Excluded config sets")
		}

		var buildEngine builder.Builder

		// Choose the build engine based on the flag
		switch flags.Engine {
		case "buildx":
			buildEngine = &builder.BuildxBuilder{}
		// case "kaniko":
		// 	buildEngine = &builder.KanikoBuilder{}
		default:
			buildEngine = &builder.DockerBuilder{}
		}

		err := buildEngine.Init()
		util.FailOnError(err, "Failed to initialize builder.")
		buildEngine.SetThreads(flags.Threads)
		buildEngine.SetDryRun(!flags.Build)

		combinations := generateVariableCombinations(img.Variables)
		for _, rawConfigSet := range combinations {
			log.Info().Str("image", name).Msg("Building")
			configSet, err := generateConfigSet(name, cfg, rawConfigSet, flags)
			if err != nil {
				log.Error().Err(err).Msg("Failed to generate config set")
				return err
			}

			// skip excluded config sets
			if isExcluded(configSet, img.Excludes) {
				log.Debug().Interface("config set", configSet).Interface("excludes", img.Excludes).Msg("Skipping excluded")
				continue
			}

			// Collect all required data
			tags := collectTags(img, configSet, name)

			// Collect labels, starting with global labels, then oci, then per image
			labels := collectOCILabels(configSet)
			templatedLabels, err := collectLabels(configSet)
			if err != nil {
				return err
			}
			maps.Copy(labels, templatedLabels)
			configSet["labels"] = labels

			// Collect build args
			buildArgs, err := collectBuildArgs(configSet)
			if err != nil {
				return err
			}
			configSet["args"] = buildArgs

			var dockerfile string
			if strings.HasSuffix(dockerfileTemplate, ".tpl") {
				dockerfile = generateDockerfilePath(dockerfileTemplate, name, configSet)
				log.Debug().Str("dockerfile", dockerfile).Msg("Generating temporary")

				// Template Dockerfile
				if err := templateFile(dockerfileTemplate, dockerfile, configSet); err != nil {
					log.Error().Err(err).Str("dockerfile", dockerfile).Msg("Failed to template Dockerfile")
					return err
				}

				// Cleanup temporary files
				if flags.Delete {
					defer util.RemoveFile(dockerfile)
				}
			} else {
				dockerfile = dockerfileTemplate
			}

			// name is required to avoid collisions between images or
			// when variables are not defined to have actual image name
			// ERROR: invalid tag "timezone-UTC": repository name must be lowercase
			currentImage := strings.ToLower(strings.Trim(fmt.Sprintf("%s-%s", name, generateCombinationString(configSet)), "-"))

			// collect building image commands
			// FIXME: I should pass templated labels here, maybe I should update configSet
			buildEngine.Build(dockerfile, currentImage, configSet, filepath.Dir(dockerfileTemplate), flags.Verbose)

			// collect tagging commands to keep order
			for _, t := range tags {
				taggedImg := generateImageName(cfg.Registry, cfg.Prefix, t)
				buildEngine.Tag(currentImage, taggedImg, flags.Verbose)
				buildEngine.Push(taggedImg, flags.Verbose)
			}

			// remove temporary tags
			buildEngine.Remove(currentImage, flags.Verbose)
		}

		if flags.Build {
			err := buildEngine.RunBuilding()
			util.FailOnError(err, "Building failed with error, check error above. Exiting.")
		}

		// let squash it
		if flags.Build && flags.Squash {
			// inspect requires images to be already built, so I need another loop here
			for _, configSet := range combinations {
				currentImage := strings.ToLower(strings.Trim(fmt.Sprintf("%s-%s", name, generateCombinationString(configSet)), "-"))
				buildEngine.Squash(currentImage, flags.Verbose)
			}
			err := buildEngine.RunSquashing()
			util.FailOnError(err, "Squashing failed with error, check error above. Exiting.")
		}

		// continue typical build
		if flags.Build {
			err := buildEngine.RunTagging()
			util.FailOnError(err, "Tagging failed with error, check error above. Exiting.")
			err = buildEngine.RunCleanup()
			util.FailOnError(err, "Dropping temporary images failed. Exiting.")
		}
		if flags.Push {
			err := buildEngine.RunPushing()
			util.FailOnError(err, "Pushing images failed, check error above. Exiting.")
		}

		// Shutdown the builder
		err = buildEngine.Shutdown()
		util.FailOnError(err, "Failed to shutdown builder.")
		fmt.Println("")
	}
	return nil
}

func generateConfigSet(imageName string, cfg *config.Config, currentConfigSet map[string]interface{}, flag config.Flags) (map[string]interface{}, error) {
	newConfigSet := make(map[string]interface{})

	// first populate global values
	newConfigSet["registry"] = cfg.Registry
	newConfigSet["prefix"] = cfg.Prefix
	newConfigSet["maintainer"] = cfg.Maintainer
	newConfigSet["labels"] = map[string]string{}
	newConfigSet["platforms"] = []string{}
	maps.Copy(newConfigSet["labels"].(map[string]string), cfg.GlobalLabels)
	newConfigSet["platforms"] = cfg.GlobalPlatforms

	// then populate image specific values
	newConfigSet["image"] = imageName
	maps.Copy(newConfigSet["labels"].(map[string]string), cfg.Images[imageName].Labels)
	if len(cfg.Images[imageName].Platforms) > 0 {
		newConfigSet["platforms"] = cfg.Images[imageName].Platforms
	}

	// check if users don't try to override reserved keys
	for k := range currentConfigSet {
		if isIgnoredKey(k) {
			return nil, fmt.Errorf("variable key '%s' is reserved and cannot be used as variable", k)
		}
	}
	maps.Copy(newConfigSet, currentConfigSet)

	// populate flag specific values
	newConfigSet["tag"] = flag.Tag

	// validate if only allowed platforms are used
	for _, p := range newConfigSet["platforms"].([]string) {
		if !isAllowedPlatform(p) {
			return nil, fmt.Errorf("platform '%s' is not allowed", p)
		}
	}

	log.Debug().Interface("config set", newConfigSet).Msg("Generated")
	return newConfigSet, nil
}

func collectLabels(configSet map[string]interface{}) (map[string]string, error) {
	labels, err := templateLabels(configSet["labels"].(map[string]string), configSet)
	if err != nil {
		return nil, err
	}
	if len(labels) > 0 {
		log.Info().Interface("labels", labels).Msg("Generating")
	}
	return labels, nil
}

func collectBuildArgs(configSet map[string]interface{}) (map[string]string, error) {
	buildArgs, err := templateLabels(configSet["args"].(map[string]string), configSet)
	if err != nil {
		return nil, err
	}
	if len(buildArgs) > 0 {
		log.Info().Interface("buildArgs", buildArgs).Msg("Generating")
	}
	return buildArgs, nil
}

func collectTags(img config.ImageConfig, configSet map[string]interface{}, name string) []string {
	tags, err := templateTags(img.Tags, configSet)
	// FIXME: return this error further
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
func generateVariableCombinations(variables map[string]interface{}) []map[string]interface{} {
	var combinations []map[string]interface{}

	// Helper function to recursively generate combinations
	var generate func(map[string]interface{}, map[string]interface{}, []string)
	generate = func(current map[string]interface{}, remaining map[string]interface{}, keys []string) {
		if len(keys) == 0 {
			combo := make(map[string]interface{})
			for k, v := range current {
				combo[k] = v
			}
			combinations = append(combinations, combo)
			return
		}

		key := keys[0]
		value := remaining[key]

		switch v := value.(type) {
		case []interface{}:
			for _, item := range v {
				current[key] = item
				generate(current, remaining, keys[1:])
			}
		case string:
			current[key] = v
			generate(current, remaining, keys[1:])
		case map[string]interface{}:
			for subKey, subValue := range v {
				current[key] = map[string]interface{}{subKey: subValue}
				generate(current, remaining, keys[1:])
			}
		default:
			current[key] = v
			generate(current, remaining, keys[1:])
		}
	}

	generate(map[string]interface{}{}, variables, getKeys(variables))
	return combinations
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func templateTags(tagTemplates []string, configSet map[string]interface{}) ([]string, error) {
	var tags []string

	for _, label := range tagTemplates {
		templated, err := templateString(label, configSet)
		if err != nil {
			return nil, err
		}
		tags = append(tags, strings.Trim(templated, " \n"))
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
		templatedLabel = strings.Trim(templatedLabel, " \n")
		templatedValue = strings.Trim(templatedValue, " \n")
		labels[templatedLabel] = templatedValue
	}

	return labels, nil
}

func sanitizeForFileName(input string) string {
	// Replace any character that is not a letter, number, or safe symbol (-, _) with an underscore
	// FIXME: This can actually result in collisions if someones uses a lot of symbols in variables
	// 		  But I didn't face it yet, maybe it's not a problem at all
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_\.]+`)
	return reg.ReplaceAllString(input, "_")
}

// to avoid tags like this
// ERROR: invalid tag "test-case-7-alpine-3.21-crazy-map_key2_value2_-timezone-utc": invalid reference format
func sanitizeForTag(input string) string {
	// Replace any character that is not a letter, number, or safe symbol (-, _) with an underscore
	// FIXME: This can actually result in collisions if someones uses a lot of symbols in variables
	// 		  But I didn't face it yet, maybe it's not a problem at all
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_\.]+`)
	return strings.Trim(reg.ReplaceAllString(input, "-"), "-")
}

func generateCombinationString(configSet map[string]interface{}) string {
	var parts []string
	for k, v := range configSet {
		if !isIgnoredKey(k) {
			// Apply sanitization to both key and value
			safeKey := sanitizeForTag(k)
			safeValue := sanitizeForTag(fmt.Sprintf("%v", v))
			parts = append(parts, fmt.Sprintf("%s-%s", safeKey, safeValue))
			log.Debug().Str("key", safeKey).Str("value", safeValue).Msg("Combining")
		}
	}
	sort.Strings(parts)
	return strings.Trim(strings.Join(parts, "-"), "-")
}

func generateDockerfilePath(dockerFileTemplate string, image string, configSet map[string]interface{}) string {
	dirname := filepath.Dir(dockerFileTemplate)
	filename := strings.Trim(fmt.Sprintf("%s-%s.Dockerfile", image, generateCombinationString(configSet)), "-")
	return filepath.Join(dirname, sanitizeForFileName(filename))
}

func generateImageName(registry string, prefix string, name string) string {
	return strings.ToLower(path.Join(registry, prefix, name))
}

func isIgnoredKey(key string) bool {
	switch key {
	case
		"image",
		"registry",
		"prefix",
		"maintainer",
		"tag",
		"labels",
		"platforms":
		return true
	}
	return false
}

func isAllowedPlatform(platform string) bool {
	switch platform {
	case
		// following https://github.com/tonistiigi/binfmt
		"linux/amd64",
		"linux/arm64",
		"linux/riscv64",
		"linux/ppc64le",
		"linux/s390x",
		"linux/386",
		"linux/arm/v7",
		"linux/arm/v6":
		return true
	}
	return false
}

func isExcluded(configSet map[string]interface{}, excludedSets []map[string]interface{}) bool {
	for _, exclusion := range excludedSets {
		matchCounter := 0
		// verify and count matching exclusion variables
		for k, v := range exclusion {
			if configSet[k] == v {
				matchCounter += 1
			}
		}
		if matchCounter == len(exclusion) { // if all conditions match
			// then exclusion condition is met
			return true
		}
	}
	return false
}
