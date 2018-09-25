package proxy

import (
	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/common/rlp"
	"github.com/paradigm-network/paradigm/types"
	"github.com/paradigm-network/paradigm/state"
	"github.com/paradigm-network/paradigm/storage"
	"github.com/rs/zerolog/log"
	"math/big"
)

// write ahead state, updated with each AppendTx
// and reset on Commit
type WriteAheadState struct {
	db      storage.Store
	stateDB *state.StateDB

	txIndex      int
	transactions []*types.Transaction
	receipts     []*types.Receipt
	allLogs      []*types.Log

	totalUsedGas *big.Int
	gp           *GasPool
}

func (was *WriteAheadState) Commit() (common.Hash, error) {
	//commit all state changes to the database
	hashArray, err := was.stateDB.CommitTo(was.db, true)
	if err != nil {
		log.Error().Err(err).Msg("Committing state")
		return common.Hash{}, err
	}
	if err := was.writeHead(); err != nil {
		log.Error().Err(err).Msg("Writing head")
		return common.Hash{}, err
	}
	if err := was.writeTransactions(); err != nil {
		log.Error().Err(err).Msg("Writing txs")
		return common.Hash{}, err
	}
	if err := was.writeReceipts(); err != nil {
		log.Error().Err(err).Msg("Writing receipts")
		return common.Hash{}, err
	}
	return hashArray, nil
}

func (was *WriteAheadState) writeHead() error {
	head := &types.Transaction{}
	if len(was.transactions) > 0 {
		head = was.transactions[len(was.transactions)-1]
	}
	return was.db.Put(headTxKey, head.Hash().Bytes())
}

func (was *WriteAheadState) writeTransactions() error {
	for _, tx := range was.transactions {
		data, err := rlp.EncodeToBytes(tx)
		if err != nil {
			return err
		}
		if err := was.db.Put(tx.Hash().Bytes(), data); err != nil {
			return err
		}
	}

	return nil
}

func (was *WriteAheadState) writeReceipts() error {
	for _, receipt := range was.receipts {
		storageReceipt := (*types.ReceiptForStorage)(receipt)
		data, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			return err
		}
		if err := was.db.Put(append(receiptsPrefix, receipt.TxHash.Bytes()...), data); err != nil {
			return err
		}
	}
	return nil
}
