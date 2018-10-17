package types

import (
	"io"

	comm "github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/p2pserver/common"
)

type Pong struct {
	Height uint64
}

//Serialize message payload
func (this Pong) Serialization(sink *comm.ZeroCopySink) error {
	sink.WriteUint64(this.Height)
	return nil
}

func (this Pong) CmdType() string {
	return common.PONG_TYPE
}

//Deserialize message payload
func (this *Pong) Deserialization(source *comm.ZeroCopySource) error {
	var eof bool
	this.Height, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}
