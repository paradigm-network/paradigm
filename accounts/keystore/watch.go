package keystore

import "github.com/rjeczalik/notify"

type watcher struct {
	ac       *accountCache
	starting bool
	running  bool
	ev       chan notify.EventInfo
	quit     chan struct{}
}

func newWatcher(ac *accountCache) *watcher {
	return &watcher{
		ac:   ac,
		ev:   make(chan notify.EventInfo, 10),
		quit: make(chan struct{}),
	}
}

func (w *watcher) close() {
	close(w.quit)
}
