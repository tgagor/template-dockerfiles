package util

import (
	"os"

	"github.com/rs/zerolog/log"
)

func RemoveFile(files ...string) {
	for _, file := range files {
		log.Debug().Str("file", file).Msg("Removing temporary")
		if err := os.Remove(file); err != nil {
			log.Error().Err(err).Str("file", file).Msg("Failed to remove")
		}
	}
}
