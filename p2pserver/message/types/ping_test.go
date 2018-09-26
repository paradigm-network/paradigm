package types

import (
	"testing"
)

func TestPingSerializationDeserialization(t *testing.T) {
	var msg Ping
	msg.Height = 1

	MessageTest(t, &msg)
}
