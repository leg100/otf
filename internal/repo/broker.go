package repo

import (
	"sync"

	"github.com/leg100/otf/internal/cloud"
)

type (
	broker struct {
		subscribers []func(event cloud.VCSEvent)
		mu          sync.RWMutex
	}
	callback func(event cloud.VCSEvent)
)

func (b *broker) Subscribe(cb callback) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers = append(b.subscribers, cb)
}

func (b *broker) publish(event cloud.VCSEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscribers {
		go sub(event)
	}
}
