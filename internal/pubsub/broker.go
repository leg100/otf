// Package pubsub implements cluster-wide publishing and subscribing of events
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/cenkalti/backoff.v1"
)

const (
	defaultChannel = "events"

	// subBufferSize is the buffer size of the channel for each subscription.
	subBufferSize = 16
)

type (
	// Broker is a pubsub Broker implemented using postgres' listen/notify
	Broker struct {
		logr.Logger

		channel       string                      // postgres notification channel name
		pool          pool                        // pool from which to acquire a dedicated connection to postgres
		islistening   chan bool                   // semaphore that's closed once broker is listening
		pid           string                      // pid uniquely identifies a broker
		registrations map[string]otf.Getter       // map of event payload type to a fn that can retrieve the payload
		subs          map[string]chan otf.Event   // subscriptions
		metrics       map[string]prometheus.Gauge // metric for each subscription

		mu sync.Mutex // sync access to maps
	}

	pool interface {
		Acquire(ctx context.Context) (*pgxpool.Conn, error)
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	}

	// pgevent is an event sent via a postgres channel
	pgevent struct {
		PayloadType string        `json:"payload_type"` // event payload type
		Event       otf.EventType `json:"event"`        // event type
		ID          string        `json:"id"`           // event payload ID
		PID         string        `json:"pid"`          // process ID that sent this event
	}
)

func NewBroker(logger logr.Logger, db otf.DB) *Broker {
	return &Broker{
		Logger:        logger.WithValues("component", "broker"),
		pid:           uuid.NewString(),
		pool:          db,
		islistening:   make(chan bool),
		channel:       defaultChannel,
		registrations: make(map[string]otf.Getter),
		subs:          make(map[string]chan otf.Event),
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

// Publish sends an event to subscribers, via postgres to subscribers on
// other machines, and via the local broker to subscribers within the same
// process.
func (b *Broker) Publish(event otf.Event) {
	b.localPublish(event)

	if event.Local {
		return
	}

	if err := b.remotePublish(event); err != nil {
		b.Error(err, "publishing event via postgres", "event", event.Type)
	}
}

// Subscribe subscribes the caller to a stream of events. Prefix is an
// identifier prefixed to a random string to helpfully identify the subscriber
// in metrics.
func (b *Broker) Subscribe(ctx context.Context, prefix string) (<-chan otf.Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	name := prefix + otf.GenerateRandomString(4)

	sub := make(chan otf.Event, subBufferSize)
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

// Register a means of reassembling a postgres message back into an otf event
func (b *Broker) Register(t reflect.Type, getter otf.Getter) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.registrations[t.String()] = getter
}

// localPublish publishes an event to subscribers on the local node
func (b *Broker) localPublish(event otf.Event) {
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

// remotePublish publishes an event to postgres for relaying onto to remote
// subscribers
func (b *Broker) remotePublish(event otf.Event) error {
	// marshal an otf event into a JSON-encoded postgres event
	id, hasID := otf.GetID(event.Payload)
	if !hasID {
		return fmt.Errorf("event payload does not have an ID field")
	}
	encoded, err := json.Marshal(&pgevent{
		PayloadType: reflect.TypeOf(event.Payload).String(),
		Event:       event.Type,
		ID:          id,
		PID:         b.pid,
	})
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("select pg_notify('%s', $1)", b.channel)
	if _, err = b.pool.Exec(context.Background(), sql, encoded); err != nil {
		return err
	}
	return nil
}

// receive handles notifications from postgres
func (b *Broker) receive(ctx context.Context, notification *pgconn.Notification) error {
	var event pgevent
	if err := json.Unmarshal([]byte(notification.Payload), &event); err != nil {
		return err
	}

	// skip notifications that this process sent.
	if event.PID == b.pid {
		return nil
	}

	getter, ok := b.registrations[event.PayloadType]
	if !ok {
		return fmt.Errorf("unregistered payload type: %s", event.PayloadType)
	}
	payload, err := getter.GetByID(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("retrieving resource: %w", err)
	}

	b.localPublish(otf.Event{
		Type:    event.Event,
		Payload: payload,
	})

	return nil
}
