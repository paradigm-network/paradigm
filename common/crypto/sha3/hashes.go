package sha3

import "hash"


//state是hash.Hash的实现类，需要实现他的6个方法才不会报红
// NewKeccak256 creates a new Keccak-256 hash.
func NewKeccak256() hash.Hash { return &state{rate: 136, outputLen: 32, dsbyte: 0x01} }

