package proxy

import (
	"github.com/paradigm-network/paradigm/common"
	"math/big"
)
type Message interface {
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address

	GasPrice() *big.Int
	Gas() *big.Int
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte
}

type TxMessage struct {
	to                      *common.Address
	from                    common.Address
	nonce                   uint64
	amount, price, gasLimit *big.Int
	data                    []byte
	checkNonce              bool
}

func NewTxMessage(from common.Address, to *common.Address, nonce uint64, amount, gasLimit, price *big.Int, data []byte, checkNonce bool) TxMessage {
	return TxMessage{
		from:       from,
		to:         to,
		nonce:      nonce,
		amount:     amount,
		price:      price,
		gasLimit:   gasLimit,
		data:       data,
		checkNonce: checkNonce,
	}
}

func (m TxMessage) From() common.Address { return m.from }
func (m TxMessage) To() *common.Address  { return m.to }
func (m TxMessage) GasPrice() *big.Int   { return m.price }
func (m TxMessage) Value() *big.Int      { return m.amount }
func (m TxMessage) Gas() *big.Int        { return m.gasLimit }
func (m TxMessage) Nonce() uint64        { return m.nonce }
func (m TxMessage) Data() []byte         { return m.data }
func (m TxMessage) CheckNonce() bool     { return m.checkNonce }
