package types

import (
	"testing"

	cm "github.com/paradigm-network/paradigm/common"
)

func Uint256ParseFromBytes(f []byte) cm.Uint256 {
	if len(f) != 32 {
		return cm.Uint256{}
	}

	var hash [32]uint8
	for i := 0; i < 32; i++ {
		hash[i] = f[i]
	}
	return cm.Uint256(hash)
}

func TestNotFoundSerializationDeserialization(t *testing.T) {
	var msg NotFound
	str := "123456"
	hash := []byte(str)
	msg.Hash = Uint256ParseFromBytes(hash)

	MessageTest(t, &msg)
}
