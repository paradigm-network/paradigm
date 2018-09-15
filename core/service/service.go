package service

import (
	"net/http"
	"encoding/json"
	"strconv"
	"github.com/paradigm-network/paradigm/core"
	"github.com/rs/zerolog"
	"github.com/paradigm-network/paradigm/common/log"
	"fmt"
)

type Service struct {
	bindAddress string
	node        *core.Node
	logger      *zerolog.Logger
}

func NewService(bindAddress string, node *core.Node) *Service {
	service := Service{
		bindAddress: bindAddress,
		node:        node,
		logger:      log.GetLogger("service"),
	}

	return &service
}

func (s *Service) Serve() {
	http.HandleFunc("/stats", s.GetStats)
	http.HandleFunc("/block/", s.GetBlock)
	err := http.ListenAndServe(s.bindAddress, nil)
	fmt.Println("Serve..")
	if err != nil {
		s.logger.Error().Err(err).Msg("Service serve error.")
	}
}

func (s *Service) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := s.node.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Service) GetBlock(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Path[len("/block/"):]
	blockIndex, err := strconv.Atoi(param)
	if err != nil {
		s.logger.Error().Err(err).Msg("Parsing block_index parameter.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	block, err := s.node.GetBlock(blockIndex)
	if err != nil {
		s.logger.Error().Err(err).Msg("Retrieving block.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}
