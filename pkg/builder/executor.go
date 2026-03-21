package builder

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/parser"
)

// ExecutePlan orchestrates the build process across all layers of the dependency graph.
func ExecutePlan(plan *parser.Plan, b Builder, flags *config.Flags) error {
	layers := plan.Layers()
	for i, layer := range layers {
		if len(layer) == 0 {
			continue
		}

		log.Info().Int("layer", i).Int("nodes", len(layer)).Msg("Executing build layer")

		if err := b.Init(); err != nil {
			log.Error().Err(err).Msg("Failed to initialize builder.")
			return err
		}
		b.SetFlags(flags)

		for _, node := range layer {
			img := node.Image
			log.Info().Str("image", img.Name).Interface("config set", img.Representation()).Msg("Processing")
			b.Queue(img)
		}

		// execute the build queue
		if err := b.Run(); err != nil {
			log.Error().Err(err).Msg("Building failed with error, check error above. Exiting.")
			return err
		}

		// Shutdown the builder
		if err := b.Terminate(); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown builder.")
			return err
		}

		fmt.Println("")
	}
	return nil
}
