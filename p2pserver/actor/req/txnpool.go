package req

import (
	"time"

	"github.com/paradigm-network/paradigm/network/actor"
	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/common/log"
	"github.com/paradigm-network/paradigm/core/types"
	"github.com/paradigm-network/paradigm/errors"
	p2pcommon "github.com/paradigm-network/paradigm/p2pserver/common"
	tc "github.com/paradigm-network/paradigm/txnpool/common"
)

const txnPoolReqTimeout = p2pcommon.ACTOR_TIMEOUT * time.Second

var txnPoolPid *actor.PID

var logger = log.GetLogger("actor req")

func SetTxnPoolPid(txnPid *actor.PID) {
	txnPoolPid = txnPid
}

//add txn to txnpool
func AddTransaction(transaction *types.Transaction) {
	if txnPoolPid == nil {
		logger.Error().Msgf("[p2p]net_server AddTransaction(): txnpool pid is nil")
		return
	}
	txReq := &tc.TxReq{
		Tx:         transaction,
		Sender:     tc.NetSender,
		TxResultCh: nil,
	}
	txnPoolPid.Tell(txReq)
}

//get txn according to hash
func GetTransaction(hash common.Uint256) (*types.Transaction, error) {
	if txnPoolPid == nil {
		logger.Warn().Msgf("[p2p]net_server tx pool pid is nil")
		return nil, errors.NewErr("[p2p]net_server tx pool pid is nil")
	}
	future := txnPoolPid.RequestFuture(&tc.GetTxnReq{Hash: hash}, txnPoolReqTimeout)
	result, err := future.Result()
	if err != nil {
		logger.Warn().Msgf("[p2p]net_server GetTransaction error: %v\n", err)
		return nil, err
	}
	return result.(tc.GetTxnRsp).Txn, nil
}
