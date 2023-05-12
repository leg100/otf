// Package pubsub implements cluster-wide publishing and subscribing of events
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf/internal"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/cenkalti/backoff.v1"
)

const (
	defaultChannel = "events"

	// subBufferSize is the buffer size of the channel for each subscription.
	subBufferSize = 16

	// postgres table operations
	insert pgop = "INSERT"
	update pgop = "UPDATE"
	del    pgop = "DELETE"
)

type (
	// Broker is a pubsub Broker implemented using postgres' listen/notify
	Broker struct {
		logr.Logger

		channel       string                               // postgres notification channel name
		pool          pool                                 // pool from which to acquire a dedicated connection to postgres
		islistening   chan bool                            // semaphore that's closed once broker is listening
		registrations map[string]internal.EventUnmarshaler // map of event payload type to a fn that can retrieve the payload
		subs          map[string]chan internal.Event       // subscriptions
		metrics       map[string]prometheus.Gauge          // metric for each subscription

		mu sync.Mutex // sync access to maps
	}

	pool interface {
		Acquire(ctx context.Context) (*pgxpool.Conn, error)
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	}

	// pgevent is an event triggered by an operation on a postgres table
	pgevent struct {
		Table     string          `json:"table"`  // pg table on which operation occured
		Operation pgop            `json:"op"`     // pg operation
		Record    json.RawMessage `json:"record"` // pg table record affected
	}

	// pgop is a postgres table operation, e.g. INSERT, UPDATE, etc.
	pgop string
)

func NewBroker(logger logr.Logger, db pool) *Broker {
	return &Broker{
		Logger:        logger.WithValues("component", "broker"),
		pool:          db,
		islistening:   make(chan bool),
		channel:       defaultChannel,
		registrations: make(map[string]internal.EventUnmarshaler),
		subs:          make(map[string]chan internal.Event),
		metrics:       make(map[string]prometheus.Gauge),
	}
}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker. The listening channel is closed once the broker has
// started listening; from this point onwards published messages will be
// forwarded.
func (b *Broker) Start(ctx context.Context, isListening chan struct{}) error {
	conn, err := b.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("unable to acquire postgres connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "listen "+b.channel); err != nil {
		return err
	}
	close(isListening) // close semaphore to indicate broker is now listening
	b.Info("listening for events")

	op := func() error {
		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				select {
				case <-ctx.Done():
					// parent has decided to shutdown so exit without error
					return nil
				default:
					b.Error(err, "waiting for postgres notification")
					return err
				}
			}

			if err := b.receive(ctx, notification); err != nil {
				b.Error(err, "received postgres notification")
				continue
			}
		}
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, nil)
}

// Publish sends an event to subscribers.
func (b *Broker) Publish(event internal.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for name, sub := range b.subs {
		// record sub's chan size
		b.metrics[name].Set(float64(len(sub)))

		// TODO: detect full channel using 'select...default:' and if full, close
		// the channel. Subs can re-subscribe if they wish (will have to
		// re-engineer subs first to handle this accordingly).
		sub <- event
	}
}

// Subscribe subscribes the caller to a stream of events. Prefix is an
// identifier prefixed to a random string to helpfully identify the subscriber
// in metrics.
func (b *Broker) Subscribe(ctx context.Context, prefix string) (<-chan internal.Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	name := prefix + internal.GenerateRandomString(4)

	sub := make(chan internal.Event, subBufferSize)
	if _, ok := b.subs[name]; ok {
		return nil, fmt.Errorf("name already taken")
	}
	b.subs[name] = sub

	totalSubscribers.Inc()

	b.metrics[name] = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "otf",
		Subsystem:   "pub_sub",
		Name:        "queue_length",
		Help:        "Total length for queue for subscriber",
		ConstLabels: prometheus.Labels{"name": name},
	})
	if err := prometheus.Register(b.metrics[name]); err != nil {
		return nil, fmt.Errorf("registering metric for subscriber: %s: %w", name, err)
	}

	// when the context is done remove the subscriber
	go func() {
		<-ctx.Done()

		totalSubscribers.Dec()

		b.mu.Lock()
		defer b.mu.Unlock()

		close(sub)
		delete(b.subs, name)

		prometheus.Unregister(b.metrics[name])
		delete(b.metrics, name)
	}()

	return sub, nil
}

// Register a means of assembling a postgres notification into an otf event
func (b *Broker) Register(table string, getter internal.EventUnmarshaler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.registrations[table] = getter
}

// receive notifications from postgres and relay them as events onto subscribers.
func (b *Broker) receive(ctx context.Context, notification *pgconn.Notification) error {
	var event pgevent
	if err := json.Unmarshal([]byte(notification.Payload), &event); err != nil {
		return err
	}

	// Convert postgres operation into OTF event type
	var eventType internal.EventType
	switch event.Operation {
	case insert:
		eventType = internal.CreatedEvent
	case update:
		eventType = internal.UpdatedEvent
	case del:
		eventType = internal.DeletedEvent
	default:
		return fmt.Errorf("unknown pg operation: %s", event.Operation)
	}

	// Lookup unmarshaler to convert record to an OTF event
	getter, ok := b.registrations[event.Table]
	if !ok {
		return fmt.Errorf("unregistered table: %s", event.Table)
	}
	payload, err := getter.UnmarshalEvent(ctx, event.Record, eventType)
	if err != nil {
		return fmt.Errorf("retrieving resource: %w", err)
	}

	b.Publish(internal.Event{
		Type:    eventType,
		Payload: payload,
	})
	return nil
}
