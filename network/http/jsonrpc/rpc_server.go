package jsonrpc

import (
	"strconv"
	"fmt"
	"net/http"
	"github.com/paradigm-network/paradigm/common/log"
	"github.com/paradigm-network/paradigm/network/http/base/rpc"
)


func StartRPCServer() error {

	logger := log.GetLogger("startRPCServer")
	logger.Info().Interface().Msg("")


	http.HandleFunc("/", rpc.Handle)

	rpc.HandleFunc("getbestblockhash", rpc.GetBestBlockHash)
	rpc.HandleFunc("getblock", rpc.GetBlock)
	rpc.HandleFunc("getblockcount", rpc.GetBlockCount)
	rpc.HandleFunc("getblockhash", rpc.GetBlockHash)
	rpc.HandleFunc("getconnectioncount", rpc.GetConnectionCount)

	rpc.HandleFunc("getrawtransaction", rpc.GetRawTransaction)
	rpc.HandleFunc("sendrawtransaction", rpc.SendRawTransaction)
	rpc.HandleFunc("getstorage", rpc.GetStorage)
	rpc.HandleFunc("getversion", rpc.GetNodeVersion)
	rpc.HandleFunc("getnetworkid", rpc.GetNetworkId)

	rpc.HandleFunc("getcontractstate", rpc.GetContractState)
	rpc.HandleFunc("getmempooltxcount", rpc.GetMemPoolTxCount)
	rpc.HandleFunc("getmempooltxstate", rpc.GetMemPoolTxState)
	rpc.HandleFunc("getsmartcodeevent", rpc.GetSmartCodeEvent)
	rpc.HandleFunc("getblockheightbytxhash", rpc.GetBlockHeightByTxHash)

	rpc.HandleFunc("getbalance", rpc.GetBalance)
	rpc.HandleFunc("getallowance", rpc.GetAllowance)
	rpc.HandleFunc("getmerkleproof", rpc.GetMerkleProof)
	rpc.HandleFunc("getblocktxsbyheight", rpc.GetBlockTxsByHeight)
	rpc.HandleFunc("getgasprice", rpc.GetGasPrice)
	rpc.HandleFunc("getunboundong", rpc.GetUnboundOng)
	rpc.HandleFunc("getgrantong", rpc.GetGrantOng)

	err := http.ListenAndServe(":"+strconv.Itoa(int(cfg.DefConfig.Rpc.HttpJsonPort)), nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
