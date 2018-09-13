package log

import (
	"github.com/rs/zerolog"
)

var Writer *RotateWriter

func InitRotateWriter(fileBase string) {
	Writer = &RotateWriter{
		Filename:   fileBase,
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
}

func GetLogger(component string) *zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: Writer}).With().Timestamp().Logger()
	logger = logger.With().Str("component", component).Logger()
	return &logger
}
