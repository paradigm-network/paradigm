package log

import (
	"testing"
)

func TestLogger(t *testing.T) {
	//InitRotateWriter("./log.log")

	logger := GetConsoleLogger("main")
	logger.Info().Msg("log info xxx")
}
