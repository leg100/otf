package inmem

import (
	"context"
	"sync"

	"github.com/leg100/otf"
)

// subBufferSize is the buffer size of the channel for each subscription.
const subBufferSize = 16

var _ otf.PubSubService = (*PubSub)(nil)

// PubSub implements a 'pub-sub' service using go channels.
type PubSub struct {
	mu   sync.Mutex
	subs []chan otf.Event
}

func NewPubSub() *PubSub {
	return &PubSub{
		subs: make([]chan otf.Event, 0),
	}
}

// Publish relays an event to a list of subscribers
func (e *PubSub) Publish(event otf.Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, sub := range e.subs {
		// TODO: detect full channel using 'select...default:' and if full, close
		// the channel. Subs can re-subscribe if they wish (will have to
		// re-engineer subs first to handle this accordingly).
		sub <- event
	}
}

// Subscribe subscribes the caller to a stream of events.
func (e *PubSub) Subscribe(ctx context.Context) <-chan otf.Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	sub := make(chan otf.Event, subBufferSize)
	e.subs = append(e.subs, sub)

	// when the context is done remove the subscriber
	go func() {
		<-ctx.Done()

		e.mu.Lock()
		defer e.mu.Unlock()

		for i, ch := range e.subs {
			if ch == sub {
				close(ch)
				e.subs = append(e.subs[:i], e.subs[i+1:]...)
				return
			}
		}
	}()

	return sub
}
