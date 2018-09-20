package jsonrpc

import (
	"fmt"
	"net/http"
	"github.com/paradigm-network/paradigm/common/log"
	"github.com/paradigm-network/paradigm/network/http/base/rpc"
	"github.com/paradigm-network/paradigm/core"
	"github.com/paradigm-network/paradigm/network/http/service"
)

func StartRPCServer(s *service.Service) error {

	logger := log.GetLogger("StartRPCServer")
	logger.Info().Interface("startRPCServer", StartRPCServer).Msg("RPCServer start success")

	http.HandleFunc("/", rpc.Handle)

	rpc.HandleFunc("GetStatus", s.GetStats)
	rpc.HandleFunc("GetBlock", s.GetBlock)

	err := http.ListenAndServe(core.DefaultConfig().RpcJSONPort, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Service serve error.")
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}

