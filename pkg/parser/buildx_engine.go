package parser

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/builder"
	"github.com/tgagor/template-dockerfiles/pkg/config"
)

type BuildxEngine struct {
}

func (p *BuildxEngine) ExecutePlan(plan *Plan, flags *config.Flags) error {
	layers := plan.Layers()
	for i, layer := range layers {
		if len(layer) == 0 {
			continue
		}

		log.Info().Int("layer", i).Int("nodes", len(layer)).Msg("Executing build layer")

		buildEngine := &builder.BuildxBuilder{}
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
		if err := buildEngine.Run(); err != nil {
			log.Error().Err(err).Msg("Building failed with error, check error above. Exiting.")
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
