package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChecksum(t *testing.T) {
	data := []byte{1, 2, 3}
	cs := Checksum(data)

	writer := NewChecksum()
	writer.Write(data)
	checksum2 := writer.Sum(nil)
	assert.Equal(t, cs[:], checksum2)

}
