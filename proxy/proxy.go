package proxy

import "github.com/paradigm-network/paradigm/types"

type AppProxy interface {
	SubmitCh() chan []byte
	CommitBlock(block types.Block) ([]byte, error)
}
