package parser

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/tgagor/template-dockerfiles/pkg/builder"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
)

type DockerEngine struct {
}

func (p *DockerEngine) Parse(cfg *config.Config, flags *config.Flags) error {
	for _, name := range cfg.ImageOrder {
		// Build only what's provided by --image flag (single image)
		if flags.Image != "" && name != flags.Image {
			continue
		}

		imageCfg := cfg.Images[name]
		log.Debug().Str("image", name).Interface("config", imageCfg).Msg("Parsing")
		log.Debug().Interface("excludes", imageCfg.Excludes).Msg("Excluded config sets")

		buildEngine := &builder.DockerBuilder{}
		if err := buildEngine.Init(); err != nil {
			log.Error().Err(err).Msg("Failed to initialize builder.")
			return err
		}
		buildEngine.SetFlags(flags)

		combinations := GenerateVariableCombinations(imageCfg.Variables)
		for _, rawConfigSet := range combinations {
			img := image.From(name, cfg, rawConfigSet, flags)
			if err := img.Validate(); err != nil {
				return err
			}

			// skip excluded config sets
			if isExcluded(img.ConfigSet(), imageCfg.Excludes) {
				log.Warn().Interface("config set", img.Representation()).Interface("excludes", imageCfg.Excludes).Msg("Skipping excluded")
				continue
			}

			// schedule for building
			log.Info().Str("image", img.Name).Interface("config set", img.Representation()).Msg("Building")
			buildEngine.Queue(img)
		}

		// execute the build queue
		if err := buildEngine.RunBuilding(); err != nil {
			log.Error().Err(err).Msg("Building failed with error, check error above. Exiting.")
			return err
		}

		// let squash it
		if err := buildEngine.RunSquashing(); err != nil {
			log.Error().Err(err).Msg("Squashing failed with error, check error above. Exiting.")
			return err
		}

		// continue typical build
		if err := buildEngine.RunTagging(); err != nil {
			log.Error().Err(err).Msg("Tagging failed with error, check error above. Exiting.")
			return err
		}
		if err := buildEngine.RunPushing(); err != nil {
			log.Error().Err(err).Msg("Pushing images failed, check error above. Exiting.")
			return err
		}

		// Shutdown the builder
		if err := buildEngine.Terminate(); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown builder.")
			return err
		}

		fmt.Println("")
	}
	return nil
}
