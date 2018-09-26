package types

import (
	"testing"
)

func TestPongSerializationDeserialization(t *testing.T) {
	var msg Pong
	msg.Height = 1

	MessageTest(t, &msg)
}
