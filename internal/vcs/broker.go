package vcs

import (
	"sync"
)

type (
	// Broker is a brokerage for publishers and subscribers of VCS events.
	Broker struct {
		subscribers []func(event Event)
		mu          sync.RWMutex
	}

	Callback func(event Event)

	Subscriber interface {
		Subscribe(cb Callback)
	}

	Publisher interface {
		Publish(Event)
	}
)

func (b *Broker) Subscribe(cb Callback) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers = append(b.subscribers, cb)
}

func (b *Broker) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscribers {
		go sub(event)
	}
}
