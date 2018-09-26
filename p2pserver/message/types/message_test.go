package types

import (
	"bytes"
	"testing"

	"github.com/paradigm-network/paradigm/p2pserver/common"
	"github.com/stretchr/testify/assert"
)

func TestMsgHdrSerializationDeserialization(t *testing.T) {
	hdr := newMessageHeader("hdrtest", 0, common.Checksum(nil))

	buf := bytes.NewBuffer(nil)
	err := writeMessageHeader(buf, hdr)
	if err != nil {
		return
	}

	dehdr, err := readMessageHeader(buf)
	assert.Nil(t, err)

	assert.Equal(t, hdr, dehdr)

}
