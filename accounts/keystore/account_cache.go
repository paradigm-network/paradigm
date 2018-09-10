package keystore

import (
	"time"
	"sync"
	"paradigm/common"
	"paradigm/accounts"
)

// accountCache is a live index of all accounts in the keystore.
type accountCache struct {
	keydir   string
	watcher  *watcher
	mu       sync.Mutex
	all      accountsByURL
	byAddr   map[common.Address][]accounts.Account
	throttle *time.Timer
	notify   chan struct{}
	fileC    fileCache
}

type accountsByURL []accounts.Account
