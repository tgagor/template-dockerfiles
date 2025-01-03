package parser

import (
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/tgagor/template-dockerfiles/pkg/builder"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

func Run(workdir string, cfg *config.Config, flags *config.Flags) error {
	for _, name := range cfg.ImageOrder {
		// Build only what's provided by --image flag (single image)
		if flags.Image != "" && name != flags.Image {
			continue
		}

		rawImg := cfg.Images[name]
		log.Debug().Str("image", name).Interface("config", rawImg).Msg("Parsing")
		dockerfileTemplate := filepath.Join(workdir, rawImg.Dockerfile)
		log.Debug().Str("dockerfile", dockerfileTemplate).Msg("Processing")
		if rawImg.Excludes != nil {
			log.Debug().Interface("excludes", rawImg.Excludes).Msg("Excluded config sets")
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

		if err := buildEngine.Init(); err != nil {
			log.Error().Err(err).Msg("Failed to initialize builder.")
			return err
		}
		buildEngine.SetFlags(flags)

		combinations := GenerateVariableCombinations(rawImg.Variables)
		for _, rawConfigSet := range combinations {
			configSet, err := GenerateConfigSet(name, cfg, rawConfigSet, flags)
			if err != nil {
				log.Error().Err(err).Msg("Failed to generate config set")
				return err
			}

			// skip excluded config sets
			if isExcluded(configSet, rawImg.Excludes) {
				log.Warn().Interface("config set", configSet).Interface("excludes", rawImg.Excludes).Msg("Skipping excluded")
				continue
			}

			img := image.From(configSet, flags)

			log.Info().Str("image", name).Interface("config set", configSet).Msg("Building")
			// name is required to avoid collisions between images or
			// when variables are not defined to have actual image name
			// ERROR: invalid tag "timezone-UTC": repository name must be lowercase
			currentImage := strings.ToLower(strings.Trim(fmt.Sprintf("%s-%s", name, generateCombinationString(configSet)), "-"))
			img.SetName(currentImage).AddVariables(rawConfigSet)

			var dockerfile string
			if strings.HasSuffix(dockerfileTemplate, ".tpl") {
				dockerfile = generateDockerfilePath(dockerfileTemplate, name, configSet)
				log.Debug().Str("dockerfile", dockerfile).Msg("Generating temporary")

				// Template Dockerfile
				if err := TemplateFile(dockerfileTemplate, dockerfile, configSet); err != nil {
					log.Error().Err(err).Str("dockerfile", dockerfile).Msg("Failed to template Dockerfile")
					return err
				}

				// Cleanup temporary dockerfiles
				if flags.Delete {
					defer util.RemoveFile(dockerfile)
				}
			} else {
				dockerfile = dockerfileTemplate
			}
			img.SetDockerfile(dockerfile).SetBuildContextDir(filepath.Dir(dockerfileTemplate))

			// schedule for building
			buildEngine.Queue(img)
		}

		// execute the build queue
		if err := buildEngine.Run(); err != nil {
			log.Error().Err(err).Msg("Building failed with error, check error above. Exiting.")
			return err
		}

		// Shutdown the builder
		if err := buildEngine.Shutdown(); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown builder.")
			return err
		}

		fmt.Println("")
	}
	return nil
}

func GenerateConfigSet(imageName string, cfg *config.Config, currentConfigSet map[string]interface{}, flag *config.Flags) (map[string]interface{}, error) {
	newConfigSet := make(map[string]interface{})

	// first populate global values
	newConfigSet["registry"] = cfg.Registry
	newConfigSet["prefix"] = cfg.Prefix
	newConfigSet["maintainer"] = cfg.Maintainer
	newConfigSet["platforms"] = cfg.GlobalPlatforms
	newConfigSet["labels"] = map[string]string{}
	newConfigSet["args"] = map[string]string{}

	// then populate image specific values
	newConfigSet["image"] = imageName
	if len(cfg.Images[imageName].Platforms) > 0 {
		newConfigSet["platforms"] = cfg.Images[imageName].Platforms
	}
	if len(newConfigSet["platforms"].([]string)) > 0 && flag.Engine != "buildx" && flag.Build {
		return nil, fmt.Errorf("engine '%s' do not support multi-platform builds, use 'buildx' instead", flag.Engine)
	}

	// check if users don't try to override reserved keys
	for k := range currentConfigSet {
		if isReservedKey(k) {
			return nil, fmt.Errorf("variable key '%s' is reserved and cannot be used as variable", k)
		}
	}

	// validate if only allowed platforms are used
	for _, p := range newConfigSet["platforms"].([]string) {
		if !isAllowedPlatform(p) {
			return nil, fmt.Errorf("platform '%s' is not allowed", p)
		}
	}

	// merge global variables with current set of variables
	maps.Copy(newConfigSet, currentConfigSet)

	// populate flag specific values
	newConfigSet["tag"] = flag.Tag

	// Collect all required data
	if tags, err := TemplateList(cfg.Images[imageName].Tags, newConfigSet); err != nil {
		return nil, err
	} else if len(tags) < 1 {
		log.Error().Str("image", imageName).Msg("No 'tags' defined for")
		return nil, fmt.Errorf("building without 'tags', would just overwrite images in place, which is pointless - add 'tags' block to continue")
	} else {
		newConfigSet["tags"] = tags
	}

	// Collect labels, starting with global labels, then oci, then per image
	labels := map[string]string{}
	maps.Copy(labels, cfg.GlobalLabels)
	maps.Copy(labels, collectOCILabels(newConfigSet))
	if templatedLabels, err := TemplateMap(cfg.Images[imageName].Labels, newConfigSet); err != nil {
		return nil, err
	} else {
		maps.Copy(labels, templatedLabels)
	}
	newConfigSet["labels"] = labels

	// Collect build args
	if buildArgs, err := TemplateMap(cfg.Images[imageName].Args, newConfigSet); err != nil {
		return nil, err
	} else {
		maps.Copy(newConfigSet["args"].(map[string]string), buildArgs)
	}

	log.Debug().Interface("config set", newConfigSet).Msg("Generated")
	return newConfigSet, nil
}

// generates all combinations of variables
func GenerateVariableCombinations(variables map[string]interface{}) []map[string]interface{} {
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

func sanitizeForFileName(input string) string {
	// Replace any character that is not a letter, number, or safe symbol (-, _) with an underscore
	// FIXME: This can actually result in collisions if someone uses a lot of symbols in variables
	// 		  But I didn't face it yet, maybe it's not a problem at all
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_\.]+`)
	return reg.ReplaceAllString(input, "_")
}

// to avoid tags like this
// ERROR: invalid tag "test-case-7-alpine-3.21-crazy-map_key2_value2_-timezone-utc": invalid reference format
func sanitizeForTag(input string) string {
	// Replace any character that is not a letter, number, or safe symbol (-, _) with an underscore
	// FIXME: This can actually result in collisions if someone uses a lot of symbols in variables
	// 		  But I didn't face it yet, maybe it's not a problem at all
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_\.]+`)
	return strings.Trim(reg.ReplaceAllString(input, "-"), "-")
}

func generateCombinationString(configSet map[string]interface{}) string {
	var parts []string
	for k, v := range configSet {
		if !isReservedKey(k) {
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

func isReservedKey(key string) bool {
	switch key {
	case
		"image",
		"registry",
		"prefix",
		"maintainer",
		"tag",
		"tags",
		"labels",
		"args",
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
