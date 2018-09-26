package common

import (
	"crypto/sha256"
	"hash"
)

// checksum implement hash.Hash interface and io.Writer
type checksum struct {
	hash.Hash
}

func (self *checksum) Size() int {
	return CHECKSUM_LEN
}

func (self *checksum) Sum(b []byte) []byte {
	temp := self.Hash.Sum(nil)
	h := sha256.Sum256(temp)

	return append(b, h[:CHECKSUM_LEN]...)
}

func NewChecksum() hash.Hash {
	return &checksum{sha256.New()}
}

func Checksum(data []byte) [CHECKSUM_LEN]byte {
	var checksum [CHECKSUM_LEN]byte
	t := sha256.Sum256(data)
	s := sha256.Sum256(t[:])

	copy(checksum[:], s[:])

	return checksum
}
