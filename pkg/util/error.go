package util

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

func FailOnError(err error, msg ...string) {
	if err != nil {
		log.Error().Err(err).Msg(strings.Join(msg, " "))
		os.Exit(1)
	}
}

func WarnOnError(err error, msg ...string) {
	if err != nil {
		log.Warn().Err(err).Msg(strings.Join(msg, " "))
	}
}
