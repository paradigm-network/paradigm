package sequentia

import (
	"math/big"
	"github.com/paradigm-network/paradigm/types"
	"github.com/paradigm-network/paradigm/common/crypto"
)

type ConsensusSorter struct {
	a     []types.Comet
	r     map[int]types.RoundInfo
	cache map[int]*big.Int
}

func NewConsensusSorter(comets []types.Comet) ConsensusSorter {
	return ConsensusSorter{
		a:     comets,
		r:     make(map[int]types.RoundInfo),
		cache: make(map[int]*big.Int),
	}
}

func (b ConsensusSorter) Len() int      { return len(b.a) }
func (b ConsensusSorter) Swap(i, j int) { b.a[i], b.a[j] = b.a[j], b.a[i] }
func (b ConsensusSorter) Less(i, j int) bool {
	irr, jrr := -1, -1
	if b.a[i].RoundReceived != nil {
		irr = *b.a[i].RoundReceived
	}
	if b.a[j].RoundReceived != nil {
		jrr = *b.a[j].RoundReceived
	}
	if irr != jrr {
		return irr < jrr
	}

	if !b.a[i].ConsensusTimestamp.Equal(b.a[j].ConsensusTimestamp) {
		return b.a[i].ConsensusTimestamp.Before(b.a[j].ConsensusTimestamp)
	}

	w := b.GetPseudoRandomNumber(*b.a[i].RoundReceived)
	wsi, _, _ := crypto.DecodeSignature(b.a[i].Signature)
	wsi = wsi.Xor(wsi, w)
	wsj, _, _ := crypto.DecodeSignature(b.a[j].Signature)
	wsj = wsj.Xor(wsj, w)
	return wsi.Cmp(wsj) < 0
}
func (b ConsensusSorter) GetPseudoRandomNumber(round int) *big.Int {
	if ps, ok := b.cache[round]; ok {
		return ps
	}
	rd := b.r[round]
	ps := rd.PseudoRandomNumber()
	b.cache[round] = ps
	return ps
}
