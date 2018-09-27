package proxy

import (
	"fmt"
	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/state"
	"github.com/paradigm-network/paradigm/storage"
	"sync"
)

type MemPool struct {
	currentState *state.StateDB
	pendingState *state.ManagedState
	storage      storage.Store
	cachingDB    state.Database
	mu           sync.RWMutex
}

func NewMemPool(root common.Hash, store storage.Store) *MemPool {
	cachingDB := state.NewDatabase(store)
	current, _ := state.New(root,cachingDB)
	memPool := &MemPool{
		currentState: current,
		pendingState: state.ManageState(current),
		storage:store,
		cachingDB:cachingDB,
	}
	return memPool
}

// State returns the virtual managed state of the mem pool.
func (pool *MemPool) GetPendingNonce(address common.Address) uint64 {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	nonce := pool.pendingState.GetNonce(address)
	fmt.Printf("pending nonce = %d \n",nonce)
	pool.pendingState.SetNonce(address, nonce+1)
	return nonce
}

func (pool *MemPool) Reset(root common.Hash) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.currentState ,_ = state.New(root,pool.cachingDB)
	pool.pendingState  = state.ManageState(pool.currentState)
}
