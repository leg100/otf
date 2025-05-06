package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/sql"
)

// subBufferSize is the buffer size of the channel for each subscription.
var subBufferSize = 100

// Optionally allow user to override the buffer size.
const overrideBufferSizeEnv = "OTF_SUB_BUFFER_SIZE"

func init() {
	overrideBufferSizeString, ok := os.LookupEnv(overrideBufferSizeEnv)
	if !ok {
		return
	}
	overrideBufferSize, err := strconv.Atoi(overrideBufferSizeString)
	if err != nil {
		panic(fmt.Sprintf("expected %s to have an integer value", overrideBufferSizeEnv))
	}
	subBufferSize = overrideBufferSize
}

// ErrSubscriptionTerminated is for use by subscribers to indicate that their
// subscription has been terminated by the broker.
var ErrSubscriptionTerminated = errors.New("broker terminated the subscription")

// Broker allows clients to subscribe to OTF events.
type Broker[T any] struct {
	logr.Logger

	subs   map[chan Event[T]]struct{} // subscriptions
	mu     sync.Mutex                 // sync access to map
	getter GetterFunc[T]
	table  string
}

// GetterFunc converts a database record into a go type
type GetterFunc[T any] func(record json.RawMessage) (T, error)

// databaseListener is the upstream database events listener
type databaseListener interface {
	RegisterTable(table string, ff sql.TableFunc)
}

func NewBroker[T any](logger logr.Logger, listener databaseListener, table string) *Broker[T] {
	b := &Broker[T]{
		Logger: logger.WithValues("component", "broker"),
		subs:   make(map[chan Event[T]]struct{}),
		table:  table,
	}
	listener.RegisterTable(table, b.forward)
	return b
}

// Subscribe subscribes the caller to a stream of events. The caller can close
// the subscription by either canceling the context or calling the returned
// unsubscribe function.
func (b *Broker[T]) Subscribe(ctx context.Context) (<-chan Event[T], func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := make(chan Event[T], subBufferSize)
	b.subs[sub] = struct{}{}

	// when the context is canceled remove the subscriber
	go func() {
		<-ctx.Done()
		b.unsubscribe(sub)
	}()

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

// forward retrieves the type T uniquely identified by id and forwards it onto
// subscribers as an event together with the action.
func (b *Broker[T]) forward(action sql.Action, record json.RawMessage) {
	var event Event[T]
	if err := json.Unmarshal(record, event.Payload); err != nil {
		b.Error(err, "retrieving type for database event", "table", b.table, "action", action)
		return
	}
	switch action {
	case sql.InsertAction:
		event.Type = CreatedEvent
	case sql.UpdateAction:
		event.Type = UpdatedEvent
	case sql.DeleteAction:
		event.Type = DeletedEvent
	default:
		b.Error(nil, "unknown action", "action", action)
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

	// forceably unsubscribe full subscribers and leave it them to re-subscribe
	for _, name := range fullSubscribers {
		b.Error(nil, "unsubscribing full subscriber", "sub", name, "queue_length", subBufferSize)
		b.unsubscribe(name)
	}
}
