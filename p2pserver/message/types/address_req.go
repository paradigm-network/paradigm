package types

import (
	"github.com/paradigm-network/paradigm/common"
	comm "github.com/paradigm-network/paradigm/p2pserver/common"
)

type AddrReq struct{}

//Serialize message payload
func (this AddrReq) Serialization(sink *common.ZeroCopySink) error {
	return nil
}

func (this *AddrReq) CmdType() string {
	return comm.GetADDR_TYPE
}

//Deserialize message payload
func (this *AddrReq) Deserialization(source *common.ZeroCopySource) error {
	return nil
}
