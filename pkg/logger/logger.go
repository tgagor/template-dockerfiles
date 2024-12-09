package logger

import (
    "log"
)

type Logger struct {
    verbose bool
}

func New(verbose bool) *Logger {
    return &Logger{verbose: verbose}
}

func (l *Logger) Info(msg string) {
    log.Println("[INFO]", msg)
}

func (l *Logger) Error(msg string) {
    log.Println("[ERROR]", msg)
}

func (l *Logger) Fatal(msg string, err error) {
    log.Fatalf("[FATAL] %s: %v\n", msg, err)
}
