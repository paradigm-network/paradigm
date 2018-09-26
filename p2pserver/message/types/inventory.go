package types

import (
	"io"

	"github.com/paradigm-network/paradigm/common"
	p2pCommon "github.com/paradigm-network/paradigm/p2pserver/common"
)

var LastInvHash common.Uint256

type InvPayload struct {
	InvType common.InventoryType
	Blk     []common.Uint256
}

type Inv struct {
	P InvPayload
}

func (this Inv) invType() common.InventoryType {
	return this.P.InvType
}

func (this *Inv) CmdType() string {
	return p2pCommon.INV_TYPE
}

//Serialize message payload
func (this Inv) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint8(uint8(this.P.InvType))

	blkCnt := uint32(len(this.P.Blk))
	sink.WriteUint32(blkCnt)
	for _, hash := range this.P.Blk {
		sink.WriteHash(hash)
	}

	return nil
}

//Deserialize message payload
func (this *Inv) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	invType, eof := source.NextUint8()
	this.P.InvType = common.InventoryType(invType)
	blkCnt, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	for i := 0; i < int(blkCnt); i++ {
		hash, eof := source.NextHash()
		if eof {
			return io.ErrUnexpectedEOF
		}

		this.P.Blk = append(this.P.Blk, hash)
	}

	if blkCnt > p2pCommon.MAX_INV_BLK_CNT {
		blkCnt = p2pCommon.MAX_INV_BLK_CNT
	}
	this.P.Blk = this.P.Blk[:blkCnt]
	return nil
}
