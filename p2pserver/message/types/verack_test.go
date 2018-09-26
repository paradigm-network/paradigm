package types

import (
	"testing"
)

func TestVerackSerializationDeserialization(t *testing.T) {
	var msg VerACK
	msg.IsConsensus = false

	MessageTest(t, &msg)
}
