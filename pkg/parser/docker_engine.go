package parser

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/tgagor/template-dockerfiles/pkg/builder"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
	"github.com/tgagor/template-dockerfiles/pkg/runner"
)

type DockerEngine struct {
}

func (p *DockerEngine) Parse(cfg *config.Config, flags *config.Flags) error {
	// collect all push tasks and push at the end
	pusher := runner.New()
	pusher.Threads(flags.Threads)

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

			// skip excluded config sets
			if isExcluded(img.ConfigSet(), imageCfg.Excludes) {
				log.Warn().Interface("config set", img.Representation()).Interface("excludes", imageCfg.Excludes).Msg("Skipping excluded")
				continue
			}

			if err := img.Validate(); err != nil {
				return err
			}

			// schedule for building
			log.Info().Str("image", img.Name).Interface("config set", img.Representation()).Msg("Processing")
			buildEngine.Queue(img)
		}

		// execute the build queue
		if flags.Build {
			if err := buildEngine.RunBuilding(); err != nil {
				log.Error().Err(err).Msg("Building failed with error, check error above. Exiting.")
				return err
			}
		}
		//  else {
		// 	log.Warn().Msg("Skipping building images. Use --build flag to build images.")
		// }

		// let squash it
		if flags.Build && flags.Squash {
			if err := buildEngine.RunSquashing(); err != nil {
				log.Error().Err(err).Msg("Squashing failed with error, check error above. Exiting.")
				return err
			}
		}

		// continue typical build
		if flags.Build {
			if err := buildEngine.RunTagging(); err != nil {
				log.Error().Err(err).Msg("Tagging failed with error, check error above. Exiting.")
				return err
			}
		}
		if flags.Push {
			pusher.AddUniq(buildEngine.CollectPushTasks()...)
		}

		// Shutdown the builder
		if err := buildEngine.Terminate(); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown builder.")
			return err
		}

		fmt.Println("")
	}

	// push only if everything builds
	if flags.Push {
		log.Error().Interface("push tasks", pusher.GetTasks()).Msg("here")

		if err := pusher.Run(); err != nil {
			log.Error().Err(err).Msg("Pushing images failed, check error above. Exiting.")
			return err
		}
	}

	return nil
}
