package proxy

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/state"
	"math/big"
)

type StateTransition struct {
	gp         *GasPool
	msg        Message
	gas        uint64
	gasPrice   *big.Int
	initialGas *big.Int
	value      *big.Int
	data       []byte
	state      *state.StateDB
}

// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ProcessMessage(msg Message, gp *GasPool, statedb *state.StateDB) ([]byte, *big.Int, bool, error) {
	st := &StateTransition{
		gp:         gp,
		msg:        msg,
		gasPrice:   msg.GasPrice(),
		initialGas: new(big.Int),
		value:      msg.Value(),
		data:       msg.Data(),
		state:      statedb,
	}
	ret, _, gasUsed, failed, err := st.TransitionOnNative()
	return ret, gasUsed, failed, err
}

func (st *StateTransition) from() common.Address {
	f := st.msg.From()
	if !st.state.Exist(f) {
		st.state.CreateAccount(f)
	}
	return f
}

func (st *StateTransition) to() common.Address {
	if st.msg == nil {
		return common.Address{}
	}
	to := st.msg.To()
	if to == nil {
		return common.Address{} // contract creation
	}

	if !st.state.Exist(*to) {
		st.state.CreateAccount(*to)
	}
	return *to
}

func (st *StateTransition) useGas(amount uint64) error {
	if st.gas < amount {
		return ErrOutOfGas
	}
	st.gas -= amount

	return nil
}

func (st *StateTransition) buyGas() error {
	mgas := st.msg.Gas()
	if mgas.BitLen() > 64 {
		return ErrOutOfGas
	}
	mgval := new(big.Int).Mul(mgas, st.gasPrice)
	var (
		state  = st.state
		sender = st.from()
	)
	if state.GetBalance(sender).Cmp(mgval) < 0 {
		return ErrInsufficientBalanceForGas
	}
	if err := st.gp.SubGas(mgas); err != nil {
		return err
	}
	st.gas += mgas.Uint64()
	st.initialGas.Set(mgas)
	state.SubBalance(sender, mgval)
	return nil
}

func (st *StateTransition) preCheck() error {
	msg := st.msg
	sender := st.from()

	// Make sure this transaction's nonce is correct
	if msg.CheckNonce() {
		nonce := st.state.GetNonce(sender)
		fmt.Printf("check msg nonce = %d , state nonce = %d \n",msg.Nonce(),nonce)
		if nonce < msg.Nonce() {
			return ErrNonceTooHigh
		} else if nonce > msg.Nonce() {
			return ErrNonceTooLow
		}
	}
	return st.buyGas()
}

func (st *StateTransition) TransitionOnNative() (ret []byte, requiredGas, usedGas *big.Int, failed bool, err error) {

	if err = st.preCheck(); err != nil {
		return
	}
	msg := st.msg
	from := st.msg.From()
	to := st.to()
	//todo gas
	intrinsicGas := new(big.Int).SetUint64(100)
	if err = st.useGas(intrinsicGas.Uint64()); err != nil {
		return nil, nil, nil, false, err
	}

	// Fail if we're trying to transfer more than the available balance
	if !CanTransfer(st.state, from, msg.Value()) {
		return nil, nil, nil, false, ErrInsufficientBalance
	}

	if !st.state.Exist(st.to()) {
		st.state.CreateAccount(to)
	}
	Transfer(st.state, from, to, st.msg.Value())

	sender := st.from()
	st.state.SetNonce(sender, st.state.GetNonce(sender)+1)
	requiredGas = new(big.Int).Set(st.gasUsed())

	st.refundGas()

	//todo add gas to miner

	return ret, requiredGas, st.gasUsed(), false, err
}

func (st *StateTransition) refundGas() {
	// Return eth for remaining gas to the sender account,
	// exchanged at the original rate.
	sender := st.from() // err already checked
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(sender, remaining)

	// Apply refund counter, capped to half of the used gas.
	uhalf := remaining.Div(st.gasUsed(), common.Big2)
	refund := math.BigMin(uhalf, st.state.GetRefund())
	st.gas += refund.Uint64()

	st.state.AddBalance(sender, refund.Mul(refund, st.gasPrice))

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(new(big.Int).SetUint64(st.gas))
}

func (st *StateTransition) gasUsed() *big.Int {
	return new(big.Int).Sub(st.initialGas, new(big.Int).SetUint64(st.gas))
}

// CanTransfer checks wether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db *state.StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db *state.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}
