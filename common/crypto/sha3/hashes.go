package sha3

import "hash"


//state implements hash.Hash
// NewKeccak256 creates a new Keccak-256 hash.
func NewKeccak256() hash.Hash { return &state{rate: 136, outputLen: 32, dsbyte: 0x01} }

