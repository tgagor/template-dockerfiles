package parser

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/tgagor/template-dockerfiles/pkg/builder"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/runner"
)

type DockerEngine struct {
}

func (p *DockerEngine) ExecutePlan(plan *Plan, flags *config.Flags) error {
	// collect all push tasks and push at the end
	pusher := runner.New()
	pusher.Threads(flags.Threads)

	layers := plan.Layers()
	for i, layer := range layers {
		if len(layer) == 0 {
			continue
		}

		log.Info().Int("layer", i).Int("nodes", len(layer)).Msg("Executing build layer")

		buildEngine := &builder.DockerBuilder{}
		if err := buildEngine.Init(); err != nil {
			log.Error().Err(err).Msg("Failed to initialize builder.")
			return err
		}
		buildEngine.SetFlags(flags)

		for _, node := range layer {
			img := node.Image
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
		if err := pusher.Run(); err != nil {
			log.Error().Err(err).Msg("Pushing images failed, check error above. Exiting.")
			return err
		}
	}

	return nil
}
