package inmem

import (
	"sync"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

// SubBufferSize is the buffer size of the channel for each subscription.
const SubBufferSize = 16

var _ otf.PubSubService = (*PubSub)(nil)

type PubSub struct {
	mu   sync.Mutex
	subs map[string]*Subscription
	logr.Logger
}

func NewPubSub(logger logr.Logger) *PubSub {
	return &PubSub{
		subs:   make(map[string]*Subscription),
		Logger: logger.WithValues("component", "event_service"),
	}
}

func (e *PubSub) Publish(event otf.Event) {
	for _, sub := range e.subs {
		select {
		case sub.c <- event:
		default:
			e.unsubscribe(sub)
		}
	}
}

func (e *PubSub) Subscribe(id string) (otf.Subscription, error) {
	// Create new subscription
	sub := &Subscription{
		service: e,
		c:       make(chan otf.Event, SubBufferSize),
		id:      id,
	}

	// Add to list of user's subscriptions. Subscriptions are stored as a map
	// for each user so we can easily delete them.
	e.subs[id] = sub

	e.Info("created subscription", "subscriber", id)

	return sub, nil
}

// Unsubscribe disconnects sub from the service.
func (e *PubSub) Unsubscribe(sub *Subscription) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.unsubscribe(sub)

	e.Info("deleted subscription", "subscriber", sub.id)
}

func (e *PubSub) unsubscribe(sub *Subscription) {
	// Only close the underlying channel once. Otherwise Go will panic.
	sub.once.Do(func() {
		close(sub.c)
	})

	delete(e.subs, sub.id)
}

// Subscription represents a stream of events.
type Subscription struct {
	service *PubSub // service subscription was created from
	id      string  // Uniquely identifies subscription

	c    chan otf.Event // channel of events
	once sync.Once      // ensures c only closed once
}

// Close disconnects the subscription from the service it was created from.
func (s *Subscription) Close() error {
	s.service.Unsubscribe(s)
	return nil
}

// C returns a receive-only channel of user-related events.
func (s *Subscription) C() <-chan otf.Event {
	return s.c
}
