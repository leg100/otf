package pubsub

import (
	"context"
	"errors"
	"sync"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/sql"
)

const (
	// subBufferSize is the buffer size of the channel for each subscription.
	subBufferSize = 100
)

// ErrSubscriptionTerminated is for use by subscribers to indicate that their
// subscription has been terminated by the broker.
var ErrSubscriptionTerminated = errors.New("broker terminated the subscription")

// Broker allows clients to subscribe to OTF events.
type Broker[T any] struct {
	logr.Logger

	subs   map[chan Event[T]]struct{} // subscriptions
	mu     sync.Mutex                 // sync access to map
	getter GetterFunc[T]
}

// GetterFunc retrieves the type T using its unique id.
type GetterFunc[T any] func(ctx context.Context, id string) (T, error)

func NewBroker[T any](logger logr.Logger, listener *sql.Listener, table string, getter GetterFunc[T]) *Broker[T] {
	b := &Broker[T]{
		Logger: logger.WithValues("component", "broker"),
		subs:   make(map[chan Event[T]]struct{}),
		getter: getter,
	}
	listener.RegisterFunc(table, b.forward)
	return b
}

// Subscribe subscribes the caller to a stream of events.
func (b *Broker[T]) Subscribe() (<-chan Event[T], func()) {
	return b.subscribe()
}

// Subscribe subscribes the caller to a stream of events.
func (b *Broker[T]) SubscribeWithContext(ctx context.Context) <-chan Event[T] {
	sub, unsub := b.subscribe()
	// when the context is canceled remove the subscriber
	go func() {
		<-ctx.Done()
		unsub()
	}()
	return sub
}

func (b *Broker[T]) subscribe() (chan Event[T], func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := make(chan Event[T], subBufferSize)
	b.subs[sub] = struct{}{}

	return sub, func() { b.unsubscribe(sub) }
}

func (b *Broker[T]) unsubscribe(sub chan Event[T]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subs[sub]; !ok {
		// already unsubscribed
		return
	}
	close(sub)
	delete(b.subs, sub)
}

// GetByID retrieves the type T uniquely identified by id and forwards it onto
// subscribers as an event together with the action.
func (b *Broker[T]) forward(ctx context.Context, id string, action sql.Action) {
	var event Event[T]
	if action == sql.DeleteAction {
		event.Payload = *new(T)
	} else {
		// insert, update event
		payload, err := b.getter(ctx, id)
		if err != nil {
			// log error
			return
		}
		event.Payload = payload
	}
	switch action {
	case sql.InsertAction:
		event.Type = CreatedEvent
	case sql.UpdateAction:
		event.Type = UpdatedEvent
	case sql.DeleteAction:
		event.Type = DeletedEvent
	default:
		// log error and return
		return
	}

	var fullSubscribers []chan Event[T]

	b.mu.Lock()
	for sub := range b.subs {
		select {
		case sub <- event:
			continue
		default:
			// could not publish event to subscriber because their buffer is
			// full, so add them to a list for action below
			fullSubscribers = append(fullSubscribers, sub)
		}
	}
	b.mu.Unlock()

	// forceably unsubscribe full subscribers and let the client re-subscribe.
	for _, name := range fullSubscribers {
		b.Error(nil, "unsubscribing full subscriber", "sub", name, "queue_length", subBufferSize)
		b.unsubscribe(name)
	}
}
