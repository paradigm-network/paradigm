package actor

import (
	"github.com/paradigm-network/paradigm/network/actor/log"
)

var (
	plog = log.New(log.DebugLevel, "[ACTOR]")
)

// SetLogLevel sets the log level for the logger.
//
// SetLogLevel is safe to call concurrently
func SetLogLevel(level log.Level) {
	plog.SetLevel(level)
}
