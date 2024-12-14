package util

import (
	"log/slog"
	"os"
	"strings"
)

func FailOnError(err error, msg ...string) {
	if err != nil {
		slog.Error(strings.Join(msg, " "), "error", err)
		os.Exit(1)
	}
}

func WarnOnError(err error, msg ...string) {
	if err != nil {
		slog.Warn(strings.Join(msg, " "), "warning", err)
	}
}
