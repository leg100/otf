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

	// maximum permitted size of payload in a postgres notification:
	// https://www.postgresql.org/docs/current/sql-notify.html
	notificationMaxSize = 7999

	// subBufferSize is the buffer size of the channel for each subscription.
	subBufferSize = 16
)

type (
	// Broker is a pubsub Broker implemented using postgres' listen/notify
	Broker struct {
		logr.Logger

		channel     string    // postgres notification channel name
		pool        pool      // pool from which to acquire a dedicated connection to postgres
		islistening chan bool // semaphore that's closed once broker is listening

		subs    map[string]chan internal.Event // subscriptions
		metrics map[string]prometheus.Gauge    // metric for each subscription
		mu      sync.Mutex                     // sync access to maps

		*marshaler
	}

	pool interface {
		Acquire(ctx context.Context) (*pgxpool.Conn, error)
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	}

	// pgevent is an event embedded within a postgres notification
	pgevent struct {
		Table   string             `json:"table"`             // pg table associated with event
		Event   internal.EventType `json:"event"`             // event type
		Payload json.RawMessage    `json:"payload,omitempty"` // event payload

		// Event payload ID. Only non-nil if pgevent exceeds max size.
		ID *string `json:"id,omitempty"`
	}
)

func NewBroker(logger logr.Logger, db pool) *Broker {
	return &Broker{
		Logger:      logger.WithValues("component", "broker"),
		pool:        db,
		islistening: make(chan bool),
		channel:     defaultChannel,
		subs:        make(map[string]chan internal.Event),
		metrics:     make(map[string]prometheus.Gauge),
		marshaler:   newMarshaler(),
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
			event, err := b.unmarshal(notification.Payload)
			if err != nil {
				b.Error(err, "received postgres notification")
				continue
			}
			b.localPublish(event)
			return nil
		}
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, nil)
}

// Publish sends an event to subscribers.
func (b *Broker) Publish(event internal.Event) {
	if event.Local {
		// send event only to local subscribers
		b.localPublish(event)
		return
	}
	// send event to postgres for relay to both local and remote subscribers
	notification, err := b.marshal(event)
	if err != nil {
		b.Error(err, "publishing event via postgres", "event", event.Type)
	}
	sql := fmt.Sprintf("select pg_notify('%s', $1)", b.channel)
	if _, err = b.pool.Exec(context.Background(), sql, notification); err != nil {
		b.Error(err, "publishing event via postgres", "event", event.Type)
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

// localPublish publishes an event to subscribers on the local node
func (b *Broker) localPublish(event internal.Event) {
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
