package types

import (
	"io"

	comm "github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/p2pserver/common"
)

type VerACK struct {
	IsConsensus bool
}

//Serialize message payload
func (this *VerACK) Serialization(sink *comm.ZeroCopySink) error {
	sink.WriteBool(this.IsConsensus)
	return nil
}

func (this *VerACK) CmdType() string {
	return common.VERACK_TYPE
}

//Deserialize message payload
func (this *VerACK) Deserialization(source *comm.ZeroCopySource) error {
	var irregular, eof bool
	this.IsConsensus, irregular, eof = source.NextBool()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return comm.ErrIrregularData
	}

	return nil
}
