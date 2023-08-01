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
	Callback func(event cloud.VCSEvent)

	Subscriber interface {
		Subscribe(cb Callback)
	}
)

func (b *broker) Subscribe(cb Callback) {
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
