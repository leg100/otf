package inmem

import (
	"sync"

	"github.com/go-logr/logr"
	"github.com/leg100/ots"
)

// EventBufferSize is the buffer size of the channel for each subscription.
const EventBufferSize = 16

var _ ots.EventService = (*EventService)(nil)

type EventService struct {
	mu   sync.Mutex
	subs map[string]*Subscription
	logr.Logger
}

// NewEventService returns a new instance of EventService.
func NewEventService(logger logr.Logger) *EventService {
	return &EventService{
		subs:   make(map[string]*Subscription),
		Logger: logger.WithValues("component", "event_service"),
	}
}

func (e *EventService) Publish(event ots.Event) {
	for _, sub := range e.subs {
		select {
		case sub.c <- event:
		default:
			e.unsubscribe(sub)
		}
	}
}

func (e *EventService) Subscribe(id string) ots.Subscription {
	// Create new subscription
	sub := &Subscription{
		service: e,
		c:       make(chan ots.Event, EventBufferSize),
	}

	// Add to list of user's subscriptions.
	// Subscritions are stored as a map for each user so we can easily delete them.
	e.subs[id] = sub

	e.Info("subscription created", "subscriber", id)

	return sub
}

// Unsubscribe disconnects sub from the service.
func (e *EventService) Unsubscribe(sub *Subscription) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.unsubscribe(sub)
}

func (e *EventService) unsubscribe(sub *Subscription) {
	// Only close the underlying channel once. Otherwise Go will panic.
	sub.once.Do(func() {
		close(sub.c)
	})

	delete(e.subs, sub.id)
}

// Subscription represents a stream of events.
type Subscription struct {
	service *EventService // service subscription was created from
	id      string        // Uniquely identifies subscription

	c    chan ots.Event // channel of events
	once sync.Once      // ensures c only closed once
}

// Close disconnects the subscription from the service it was created from.
func (s *Subscription) Close() error {
	s.service.Unsubscribe(s)
	return nil
}

// C returns a receive-only channel of user-related events.
func (s *Subscription) C() <-chan ots.Event {
	return s.c
}
