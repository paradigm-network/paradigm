package types

import (
	"testing"
)

func TestAddrReqSerializationDeserialization(t *testing.T) {
	var msg AddrReq

	MessageTest(t, &msg)
}
