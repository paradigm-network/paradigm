package log

import (
	"github.com/rs/zerolog"
	"os"
	"fmt"
)

var Writer *RotateWriter

func InitRotateWriter(fileBase string) {
	if Writer == nil {
		Writer = &RotateWriter{
			Filename:   fileBase,
			MaxSize:    50, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		}
	}
}

func GetConsoleLogger(component string) *zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	logger = logger.With().Str("component", component).Logger()
	return &logger
}


func GetLogger(component string) *zerolog.Logger {
	if Writer == nil {
		fmt.Println("Please init RotateWriter first.")
		os.Exit(2)
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: Writer}).With().Timestamp().Logger()
	logger = logger.With().Str("component", component).Logger()
	return &logger
}
