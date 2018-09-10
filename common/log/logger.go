package log

import (
	"github.com/rs/zerolog"
	"os"
)

func GetLogger(component string) *zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger = logger.With().Str("component", component).Logger()
	return &logger
}
