package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/leg100/otf/internal/logr"
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
	logger logr.Logger
	// subscriptions, mapped to a channel that is closed when the subscription
	// is removed.
	subs    map[chan Event[T]]chan struct{}
	mu      sync.Mutex // sync access to map
	table   sql.Table
	enabled bool
}

func NewBroker[T any](logger logr.Logger, table sql.Table) *Broker[T] {
	b := &Broker[T]{
		logger:  logger.WithValues("component", "broker"),
		subs:    make(map[chan Event[T]]chan struct{}),
		table:   table,
		enabled: true,
	}
	return b
}

func (b *Broker[T]) Table() sql.Table { return b.table }

func (b *Broker[T]) Enable() {
	b.mu.Lock()
	b.enabled = true
	b.mu.Unlock()
}

// Disable the broker. All subscribers are forcily unsubscribed.
func (b *Broker[T]) Disable() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.enabled = false
	for sub := range b.subs {
		b.unsubscribeWithoutLock(sub)
	}
}

// Subscribe subscribes the caller to a stream of events. The caller can close
// the subscription by either canceling the context or calling the returned
// unsubscribe function.
func (b *Broker[T]) Subscribe(ctx context.Context) (<-chan Event[T], func(), error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.enabled {
		return nil, nil, fmt.Errorf("event broker is disabled")
	}

	sub := make(chan Event[T], subBufferSize)
	// unsubscribed is closed when sub is removed from the broker.
	unsubscribed := make(chan struct{})
	b.subs[sub] = unsubscribed

	// when the context is canceled remove the subscriber. The goroutine also
	// exits once the subscriber has been removed by other means - otherwise it
	// would outlive the subscription for as long as the caller's context is
	// alive, keeping sub and its buffer of subBufferSize events alive with it.
	// Callers such as the runner subscribe repeatedly with a context that lives
	// for the lifetime of the process.
	go func() {
		select {
		case <-ctx.Done():
			b.unsubscribe(sub)
		case <-unsubscribed:
		}
	}()

	return sub, func() { b.unsubscribe(sub) }, nil
}

// numSubs is the number of current subscriptions.
func (b *Broker[T]) numSubs() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return len(b.subs)
}

func (b *Broker[T]) unsubscribe(sub chan Event[T]) {
	b.mu.Lock()
	b.unsubscribeWithoutLock(sub)
	b.mu.Unlock()
}

func (b *Broker[T]) unsubscribeWithoutLock(sub chan Event[T]) {
	done, ok := b.subs[sub]
	if !ok {
		// already unsubscribed
		return
	}
	close(sub)
	delete(b.subs, sub)
	// release the goroutine watching the subscriber's context
	close(done)
}

// Forward retrieves the type T uniquely identified by id and forwards it onto
// subscribers as an event together with the action.
func (b *Broker[T]) Forward(sqlEvent sql.Event) {
	event := Event[T]{
		Time: sqlEvent.Time,
	}
	if err := json.Unmarshal(sqlEvent.Record, &event.Payload); err != nil {
		b.logger.Error(err, "unmarshaling event from database record", "action", sqlEvent.Action, "record", string(sqlEvent.Record))
		return
	}
	switch sqlEvent.Action {
	case sql.InsertAction:
		event.Type = CreatedEvent
	case sql.UpdateAction:
		event.Type = UpdatedEvent
	case sql.DeleteAction:
		event.Type = DeletedEvent
	default:
		b.logger.Error(nil, "unknown action", "action", sqlEvent.Action)
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
		b.logger.Error(nil, "unsubscribing full subscriber", "sub", name, "queue_length", subBufferSize)
		b.unsubscribe(name)
	}
}
