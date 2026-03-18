package logger

import (
	"os"
	"github.com/rs/zerolog"
)

func NewLogger(serviceName string) zerolog.Logger {
	return zerolog.New(os.Stdout).With().Timestamp().Str("service", serviceName).Logger()
}