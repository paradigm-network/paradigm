package keystore

import (
	"time"
	"sync"
	"gopkg.in/fatih/set.v0"
)

// fileCache is a cache of files seen during scan of keystore.
type fileCache struct {
	all     *set.SetNonTS // Set of all files from the keystore folder
	lastMod time.Time     // Last time instance when a file was modified
	mu      sync.RWMutex
}
