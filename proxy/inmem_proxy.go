package proxy

import (
	"github.com/paradigm-network/paradigm/common/log"
	"github.com/paradigm-network/paradigm/config"
	"github.com/paradigm-network/paradigm/storage"
	"github.com/paradigm-network/paradigm/types"
	"github.com/rs/zerolog"
)

type AppProxy interface {
	SubmitCh() chan []byte
	CommitBlock(block types.Block) ([]byte, error)
}

//InmemProxy is used for testing
type InmemAppProxy struct {
	submitCh              chan []byte
	stateHash             []byte
	committedTransactions [][]byte
	logger                *zerolog.Logger
	store                 storage.Store
	service               *Service
	state                 *State
}

func NewInmemAppProxy(config *config.Config, store storage.Store) *InmemAppProxy {
	logger := log.GetLogger("InMemProxy")
	submitCh := make(chan []byte)
	state, err := NewState(store)
	if err != nil {
		logger.Error().Err(err).Msg("Create AppProxy error")
		return nil
	}

	service := NewService(config.KeyStoreDir,
		config.SequentiaAddress,
		config.PwdFile,
		state,
		submitCh)
	proxy := &InmemAppProxy{
		stateHash:             []byte{},
		committedTransactions: [][]byte{},
		logger:                logger,
		service:               service,
		submitCh:              submitCh,
		state:                 state,
		store:                 store,
	}
	proxy.Run()
	return proxy
}

func (p *InmemAppProxy) Run() {
	p.service.Run()
}

func (iap *InmemAppProxy) commit(block types.Block) ([]byte, error) {
	stateHash, err := iap.state.ProcessBlock(block)
	return stateHash.Bytes(), err

}

//------------------------------------------------------------------------------
//Implement AppProxy Interface

func (p *InmemAppProxy) SubmitCh() chan []byte {
	return p.submitCh
}

func (p *InmemAppProxy) CommitBlock(block types.Block) (stateHash []byte, err error) {
	p.logger.Info().
		Int("round_received", block.RoundReceived()).
		Int("txs", len(block.Transactions())).
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
