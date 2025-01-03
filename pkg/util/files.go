package util

import (
	"os"
	"regexp"

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

func SanitizeForFileName(input string) string {
	// Replace any character that is not a letter, number, or safe symbol (-, _) with an underscore
	// FIXME: This can actually result in collisions if someones uses a lot of symbols in variables
	// 		  But I didn't face it yet, maybe it's not a problem at all
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_\.]+`)
	return reg.ReplaceAllString(input, "_")
}
