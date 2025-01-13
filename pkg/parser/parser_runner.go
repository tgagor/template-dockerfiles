package parser

import (
	"github.com/tgagor/template-dockerfiles/pkg/config"
)

func Run(cfg *config.Config, flags *config.Flags) error {
	var engine Engine
	switch flags.Engine {
	case "buildx":
		engine = &BuildxEngine{}
	default:
		engine = &DockerEngine{}
	}

	parser := NewParser(cfg, flags)
	parser.SetEngine(engine)
	return parser.Parse()

	// for _, name := range cfg.ImageOrder {
	// 	// Build only what's provided by --image flag (single image)
	// 	if flags.Image != "" && name != flags.Image {
	// 		continue
	// 	}

	// 	imageCfg := cfg.Images[name]
	// 	log.Debug().Str("image", name).Interface("config", imageCfg).Msg("Parsing")
	// 	log.Debug().Interface("excludes", imageCfg.Excludes).Msg("Excluded config sets")

	// 	var buildEngine builder.Builder
	// 	// Choose the build engine based on the flag
	// 	switch flags.Engine {
	// 	case "buildx":
	// 		buildEngine = &builder.BuildxBuilder{}
	// 	// case "kaniko":
	// 	// 	buildEngine = &builder.KanikoBuilder{}
	// 	// case "podman":
	// 	// 	buildEngine = &builder.PodmanBuilder{}
	// 	default:
	// 		buildEngine = &builder.DockerBuilder{}
	// 	}

	// 	if err := buildEngine.Init(); err != nil {
	// 		log.Error().Err(err).Msg("Failed to initialize builder.")
	// 		return err
	// 	}
	// 	buildEngine.SetFlags(flags)

	// 	combinations := GenerateVariableCombinations(imageCfg.Variables)
	// 	for _, rawConfigSet := range combinations {
	// 		img := image.From(name, cfg, rawConfigSet, flags)
	// 		if err := img.Validate(); err != nil {
	// 			return err
	// 		}

	// 		// skip excluded config sets
	// 		if isExcluded(img.ConfigSet(), imageCfg.Excludes) {
	// 			log.Warn().Interface("config set", img.Representation()).Interface("excludes", imageCfg.Excludes).Msg("Skipping excluded")
	// 			continue
	// 		}

	// 		// schedule for building
	// 		log.Info().Str("image", img.Name).Interface("config set", img.Representation()).Msg("Building")
	// 		buildEngine.Queue(img)
	// 	}

	// 	// execute the build queue
	// 	if err := buildEngine.Run(); err != nil {
	// 		log.Error().Err(err).Msg("Building failed with error, check error above. Exiting.")
	// 		return err
	// 	}

	// 	// Shutdown the builder
	// 	if err := buildEngine.Terminate(); err != nil {
	// 		log.Error().Err(err).Msg("Failed to shutdown builder.")
	// 		return err
	// 	}

	// 	fmt.Println("")
	// }
	// return nil
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
