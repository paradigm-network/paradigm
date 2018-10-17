package types

import (
	"io"

	comm "github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/p2pserver/common"
)

type Ping struct {
	Height uint64
}

//Serialize message payload
func (this Ping) Serialization(sink *comm.ZeroCopySink) error {
	sink.WriteUint64(this.Height)
	return nil
}

func (this *Ping) CmdType() string {
	return common.PING_TYPE
}

//Deserialize message payload
func (this *Ping) Deserialization(source *comm.ZeroCopySource) error {
	var eof bool
	this.Height, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}
