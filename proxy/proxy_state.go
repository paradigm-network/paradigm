package proxy

import (
	"bytes"
	"fmt"
	"github.com/paradigm-network/paradigm/common/log"
	"github.com/paradigm-network/paradigm/storage"
	"github.com/paradigm-network/paradigm/types"
	"github.com/rs/zerolog"
	"math/big"
	"sync"

	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/common/math"
	"github.com/paradigm-network/paradigm/common/rlp"
	"github.com/paradigm-network/paradigm/state"
)

var (
	chainID        = big.NewInt(1)
	gasLimit       = big.NewInt(1000000000000000000)
	txMetaSuffix   = []byte{0x01}
	receiptsPrefix = []byte("receipts-")
	MIPMapLevels   = []uint64{1000000, 500000, 100000, 50000, 1000}
	headTxKey      = []byte("LastTx")
)

type State struct {
	db          storage.Store
	commitMutex sync.Mutex
	statedb     *state.StateDB
	was         *WriteAheadState

	signer types.Signer

	logger *zerolog.Logger
}

func NewState(store storage.Store) (*State, error) {
	s := &State{
		db:store,
		logger:log.GetLogger("proxy_state"),
		signer:types.NewBasicSigner(),
	}
	if err := s.InitState(); err != nil {
		return nil, err
	}

	s.resetWAS()

	return s, nil
}

//------------------------------------------------------------------------------
// Call is done on a copy of the state...we dont want any changes to be persisted
// Call is a readonly operation
//	func (s *State) Call(callMsg Message) ([]byte, error) {
//	s.logger.Info().Msg("Call")
//	s.commitMutex.Lock()
//	defer s.commitMutex.Unlock()
//
//	// Apply the transaction to the current state (included in the env)
//	res, _, _, err := ProcessMessage(callMsg, s.was.gp,s.was.stateDB.Copy())
//	if err != nil {
//		s.logger.Error().Err(err).Msg("Executing Call on WAS")
//		return nil, err
//	}
//
//	return res, err
//}

func (s *State) ProcessBlock(block types.Block) (common.Hash, error) {
	fmt.Println("Process Block")
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()
	blockHashBytes, _ := block.Hash()
	blockHash := common.BytesToHash(blockHashBytes)

	for txIndex, txBytes := range block.Transactions() {
		if err := s.applyTransaction(txBytes, txIndex, blockHash); err != nil {
			return common.Hash{}, err
		}
	}

	return s.commit()
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

//applyTransaction applies a transaction to the WAS
func (s *State) applyTransaction(txBytes []byte, txIndex int, blockHash common.Hash) error {

	var t types.Transaction
	if err := rlp.Decode(bytes.NewReader(txBytes), &t); err != nil {
		s.logger.Error().Err(err).Msg("Decoding Transaction")
		return err
	}
	s.logger.Info().Str("hash", t.Hash().Hex()).Str("tx", t.String()).Msg("Decoded tx")

	msg, err := t.AsMessage(s.signer)
	if err != nil {
		s.logger.Error().Err(err).Msg("Converting Transaction to Message")
		return err
	}

	//Prepare the ethState with transaction Hash so that it can be used in emitted
	//logs
	s.was.stateDB.Prepare(t.Hash(), blockHash, txIndex)

	// Apply the transaction to the current state (included in the env)
	_, gas, failed, err := ProcessMessage(msg, s.was.gp, s.was.stateDB)
	if err != nil {
		s.logger.Error().Err(err).Msg("Applying transaction to WAS")
		return err
	}

	s.was.totalUsedGas.Add(s.was.totalUsedGas, gas)

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	root := s.was.stateDB.IntermediateRoot(true) //this has side effects. It updates StateObjects (SmartContract memory)
	receipt := types.NewReceipt(root.Bytes(), failed, s.was.totalUsedGas)
	receipt.TxHash = t.Hash()
	receipt.GasUsed = new(big.Int).Set(gas)
	// if the transaction created a contract, store the creation address in the receipt.
	//todo vm not support yet.
	//if msg.To() == nil {
	//	receipt.ContractAddress = crypto.CreateAddress(msg.From(), t.Nonce())
	//}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = s.was.stateDB.GetLogs(t.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	s.was.txIndex++
	s.was.transactions = append(s.was.transactions, &t)
	s.was.receipts = append(s.was.receipts, receipt)
	s.was.allLogs = append(s.was.allLogs, receipt.Logs...)

	s.logger.Info().Str("hash", t.Hash().Hex()).Msg("Applied tx to WAS")

	return nil
}

func (s *State) commit() (common.Hash, error) {
	//commit all state changes to the database
	root, err := s.was.Commit()
	if err != nil {
		s.logger.Error().Err(err).Msg("Committing WAS")
		return root, err
	}

	// reset the write ahead state for the next block
	// with the latest eth state
	s.statedb = s.was.stateDB
	s.logger.Info().Str("root", root.Hex()).Msg("Committed")
	s.resetWAS()

	return root, nil
}

func (s *State) resetWAS() {
	state := s.statedb.Copy()
	s.was = &WriteAheadState{
		db:           s.db,
		stateDB:      state,
		txIndex:      0,
		totalUsedGas: big.NewInt(0),
		gp:           new(GasPool).AddGas(gasLimit),
	}
	s.logger.Info().Msg("Reset Write Ahead State")
}

//------------------------------------------------------------------------------

func (s *State) InitState() error {

	rootHash := common.Hash{}
	//get head transaction hash
	headTxHash := common.Hash{}
	tx := &types.Transaction{}
	emptyTxHash := tx.Hash()
	data, _ := s.db.Get(headTxKey)
	if len(data) != 0 {
		headTxHash = common.BytesToHash(data)
		s.logger.Info().Str("head_tx", headTxHash.Hex()).Msg("Loading state from existing head")
		if headTxHash == emptyTxHash {
			// there is no tx before use root continue.
			//use root to initialise the state
			var err error
			//cache wrapped state db.
			s.statedb, err = state.New(rootHash, state.NewDatabase(s.db))
			s.logger.Info().Str("root", rootHash.Hex()).Msg("There is no tx before use root to continue initialise the state")
			return err
		}
		//get head tx receipt
		headTxReceipt, err := s.GetReceipt(headTxHash)
		if err != nil {
			s.logger.Error().Err(err).Msg("Head transaction receipt missing")
			return err
		}

		//extract root from receipt
		if len(headTxReceipt.PostState) != 0 {
			rootHash = common.BytesToHash(headTxReceipt.PostState)
			s.logger.Info().Str("root", rootHash.Hex()).Msg("Head transaction root")
		}
	}

	//use root to initialise the state
	var err error
	//cache wrapped state db.
	s.statedb, err = state.New(rootHash, state.NewDatabase(s.db))
	s.logger.Info().Str("root", rootHash.Hex()).Msg("Use root to initialise the state")
	return err
}

func (s *State) CreateAccounts(accounts AccountMap) error {
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()

	for addr, account := range accounts {
		address := common.HexToAddress(addr)
		s.was.stateDB.AddBalance(address, math.MustParseBig256(account.Balance))
		s.was.stateDB.SetCode(address, common.Hex2Bytes(account.Code))
		for key, value := range account.Storage {
			s.was.stateDB.SetState(address, common.HexToHash(key), common.HexToHash(value))
		}
		s.logger.Info().Str("address", addr).Msg("Adding account")
	}

	_, err := s.commit()

	return err
}

func (s *State) GetBalance(addr common.Address) *big.Int {
	return s.statedb.GetBalance(addr)
}

func (s *State) GetNonce(addr common.Address) uint64 {
	return s.was.stateDB.GetNonce(addr)
}

func (s *State) GetTransaction(hash common.Hash) (*types.Transaction, error) {
	// Retrieve the transaction itself from the database
	data, err := s.db.Get(hash.Bytes())
	if err != nil {
		s.logger.Error().Err(err).Msg("GetTransaction")
		return nil, err
	}
	var tx types.Transaction
	if err := rlp.DecodeBytes(data, &tx); err != nil {
		s.logger.Error().Err(err).Msg("Decoding Transaction")
		return nil, err
	}

	return &tx, nil
}

func (s *State) GetReceipt(txHash common.Hash) (*types.Receipt, error) {
	data, err := s.db.Get(append(receiptsPrefix, txHash[:]...))
	if err != nil {
		s.logger.Error().Err(err).Msg("GetReceipt")
		return nil, err
	}
	var receipt types.ReceiptForStorage
	if err := rlp.DecodeBytes(data, &receipt); err != nil {
		s.logger.Error().Err(err).Msg("Decoding Receipt")
		return nil, err
	}

	return (*types.Receipt)(&receipt), nil
}
