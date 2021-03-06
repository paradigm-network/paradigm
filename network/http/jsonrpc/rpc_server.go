package jsonrpc

import (
	"fmt"
	"github.com/paradigm-network/paradigm/config"
	"net/http"
	"github.com/paradigm-network/paradigm/common/log"
	"github.com/paradigm-network/paradigm/network/http/base/rpc"
	"github.com/paradigm-network/paradigm/network/http/service"
)

func StartRPCServer(conf *config.Config, s *service.Service) error {

	logger := log.GetLogger("jsonrpc")
	logger.Info().Msg("RPCServer starting")

	http.HandleFunc("/", rpc.Handle)

	rpc.HandleFunc("GetStatus", s.GetStats)
	rpc.HandleFunc("GetBlock", s.GetBlock)

	err := http.ListenAndServe(conf.RpcAddr, nil)
	if err != nil {
		logger.Error().Msg("Service serve error.")
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}























