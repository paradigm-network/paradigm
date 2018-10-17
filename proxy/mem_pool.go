package proxy

import (
	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/state"
	"github.com/paradigm-network/paradigm/types"
	"github.com/rs/zerolog/log"
	"math/big"
)

type TxPool struct {
	stateDB      *state.StateDB
	signer       types.Signer
	gasLimit     *big.Int
	totalUsedGas *big.Int
	gp           *GasPool
}

func NewTxPool(state *state.StateDB,
	signer types.Signer,
	gasLimit *big.Int) *TxPool {

	return &TxPool{
		stateDB:  state,
		signer:   signer,
		gasLimit: gasLimit,
	}
}

func (p *TxPool) Reset(root common.Hash) error {

	err := p.stateDB.Reset(root)
	if err != nil {
		return err
	}

	p.totalUsedGas = big.NewInt(0)
	p.gp = new(GasPool).AddGas(p.gasLimit)

	return nil
}

func (p *TxPool) CheckTx(tx *types.Transaction) error {

	msg, err := tx.AsMessage(p.signer)
	if err != nil {
		log.Error().Err(err).Msg("Converting Transaction to Message")
		return err
	}

	_, gas, _, err := ProcessMessage(msg,p.gp,p.stateDB)

	p.totalUsedGas.Add(p.totalUsedGas, gas)

	return nil
}

func (p *TxPool) GetNonce(addr common.Address) uint64 {
	return p.stateDB.GetNonce(addr)
}