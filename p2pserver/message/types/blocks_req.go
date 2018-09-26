package types

import (
	"io"

	comm "github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/p2pserver/common"
)

type BlocksReq struct {
	HeaderHashCount uint8
	HashStart       comm.Uint256
	HashStop        comm.Uint256
}

//Serialize message payload
func (this *BlocksReq) Serialization(sink *comm.ZeroCopySink) error {
	sink.WriteUint8(this.HeaderHashCount)
	sink.WriteHash(this.HashStart)
	sink.WriteHash(this.HashStop)

	return nil
}

func (this *BlocksReq) CmdType() string {
	return common.GET_BLOCKS_TYPE
}

//Deserialize message payload
func (this *BlocksReq) Deserialization(source *comm.ZeroCopySource) error {
	var eof bool
	this.HeaderHashCount, eof = source.NextUint8()
	this.HashStart, eof = source.NextHash()
	this.HashStop, eof = source.NextHash()

	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
