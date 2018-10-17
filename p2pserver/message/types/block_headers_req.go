package types

import (
	"io"

	"github.com/paradigm-network/paradigm/common"
	comm "github.com/paradigm-network/paradigm/p2pserver/common"
)

type HeadersReq struct {
	Len       uint8
	HashStart common.Uint256
	HashEnd   common.Uint256
}

//Serialize message payload
func (this *HeadersReq) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint8(this.Len)
	sink.WriteHash(this.HashStart)
	sink.WriteHash(this.HashEnd)
	return nil
}

func (this *HeadersReq) CmdType() string {
	return comm.GET_HEADERS_TYPE
}

//Deserialize message payload
func (this *HeadersReq) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Len, eof = source.NextUint8()
	this.HashStart, eof = source.NextHash()
	this.HashEnd, eof = source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}
