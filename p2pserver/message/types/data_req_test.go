package types

import (
	"testing"

	cm "github.com/paradigm-network/paradigm/common"
)

func TestDataReqSerializationDeserialization(t *testing.T) {
	var msg DataReq
	msg.DataType = 0x02

	hashstr := "8932da73f52b1e22f30c609988ed1f693b6144f74fed9a2a20869afa7abfdf5e"
	bhash, _ := cm.HexToBytes(hashstr)
	copy(msg.Hash[:], bhash)

	MessageTest(t, &msg)
}
