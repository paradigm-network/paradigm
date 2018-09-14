package proxy

import (
	"github.com/paradigm-network/paradigm/types"
	"github.com/paradigm-network/paradigm/common/crypto"
	"github.com/rs/zerolog"
	"github.com/paradigm-network/paradigm/common/log"
)

//InmemProxy is used for testing
type InmemAppProxy struct {
	submitCh              chan []byte
	stateHash             []byte
	committedTransactions [][]byte
	logger                *zerolog.Logger
}

func NewInmemAppProxy() *InmemAppProxy {
	return &InmemAppProxy{
		submitCh:              make(chan []byte),
		stateHash:             []byte{},
		committedTransactions: [][]byte{},
		logger:                log.GetLogger("InMemProxy"),
	}
}

func (iap *InmemAppProxy) commit(block types.Block) ([]byte, error) {

	iap.committedTransactions = append(iap.committedTransactions, block.Transactions()...)

	hash := iap.stateHash
	for _, t := range block.Transactions() {
		tHash := crypto.SHA256(t)
		hash = crypto.SimpleHashFromTwoHashes(hash, tHash)
	}

	iap.stateHash = hash

	return iap.stateHash, nil

}

//------------------------------------------------------------------------------
//Implement AppProxy Interface

func (p *InmemAppProxy) SubmitCh() chan []byte {
	return p.submitCh
}

func (p *InmemAppProxy) CommitBlock(block types.Block) (stateHash []byte, err error) {
	p.logger.Info().
		Int("round_received",block.RoundReceived()).
		Int("txs",len(block.Transactions())).
		Msg("InMemProxy CommitBlock")
	return p.commit(block)
}

//------------------------------------------------------------------------------

func (p *InmemAppProxy) SubmitTx(tx []byte) {
	p.submitCh <- tx
}

func (p *InmemAppProxy) GetCommittedTransactions() [][]byte {
	return p.committedTransactions
}
