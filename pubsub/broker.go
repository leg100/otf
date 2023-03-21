// Package pubsub implements cluster-wide publishing and subscribing of events
package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

// a unique identity string for distinguishing this process from other otfd
// processes
var pid = uuid.NewString()

type (
	Broker interface {
		otf.PubSubService
		Register(table string, getter Getter)
		Start(context.Context) error
	}

	// broker is a pubsub broker implemented using postgres' listen/notify
	broker struct {
		logr.Logger

		channel string                      // postgres notification channel name
		pool    pool                        // pool from which to acquire a dedicated connection to postgres
		pid     string                      // each pubsub maintains a unique identifier to distriguish messages it
		tables  map[string]Getter           // map of postgres table names to getters.
		subs    map[string]chan otf.Event   // subscriptions
		metrics map[string]prometheus.Gauge // metric for each subscription

		mu sync.Mutex // sync access to maps
	}

	BrokerConfig struct {
		ChannelName *string
		PID         *string
		PoolDB      otf.DB
	}

	// Getter retrieves an OTF resource using its ID.
	Getter interface {
		GetByID(context.Context, string) (any, error)
	}

	pool interface {
		Acquire(ctx context.Context) (*pgxpool.Conn, error)
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	}

	// message is the schema of the payload for use in the postgres notification channel.
	message struct {
		// Table is the postgres table on which the event occured
		Table string `json:"relation"`
		// Action is the type of change made to the relation
		Action string `json:"action"`
		// ID is the primary key of the changed row
		ID string `json:"id"`
		// PID is the process id that sent this event
		PID string `json:"pid"`
	}
)

func NewBroker(logger logr.Logger, cfg BrokerConfig) (*broker, error) {
	// required config
	if cfg.PoolDB == nil {
		return nil, errors.New("missing database connection pool")
	}
	pool, err := cfg.PoolDB.Pool()
	if err != nil {
		return nil, err
	}

	broker := &broker{
		Logger:  logger.WithValues("component", "pubsub"),
		pid:     pid,
		pool:    pool,
		channel: defaultChannel,
		tables:  make(map[string]Getter),
		subs:    make(map[string]chan otf.Event),
		metrics: make(map[string]prometheus.Gauge),
	}

	// optional config
	if cfg.ChannelName != nil {
		broker.channel = *cfg.ChannelName
	}
	if cfg.PID != nil {
		broker.pid = *cfg.PID
	}

	return broker, nil
}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker.
func (b *broker) Start(ctx context.Context) error {
	conn, err := b.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("unable to acquire postgres connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "listen "+b.channel); err != nil {
		return err
	}

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
	return backoff.RetryNotify(op, backoff.NewExponentialBackOff(), nil)
}

// Publish sends an event to subscribers, via postgres to subscribers on
// other machines, and via the local broker to subscribers within the same
// process.
func (b *broker) Publish(event otf.Event) {
	b.localPublish(event)

	if event.Local {
		return
	}

	if err := b.remotePublish(event); err != nil {
		b.Error(err, "publishing message via postgres", "event", event.Type)
	}
}

// Subscribe subscribes the caller to a stream of events.
func (b *broker) Subscribe(ctx context.Context, name string) (<-chan otf.Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

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
		return nil, err
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
func (b *broker) Register(table string, getter Getter) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tables[table] = getter
}

// localPublish publishes an event to subscribers on the local node
func (b *broker) localPublish(event otf.Event) {
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
func (b *broker) remotePublish(event otf.Event) error {
	// marshal an otf event into a JSON-encoded postgres message
	id, hasID := otf.GetID(event.Payload)
	if !hasID {
		return fmt.Errorf("event payload does not have an ID field")
	}
	parts := strings.SplitN(string(event.Type), "_", 2)
	if len(parts) < 2 {
		return fmt.Errorf("event has an invalid type format: %s", event.Type)
	}
	msg, err := json.Marshal(&message{
		Table:  parts[0],
		Action: parts[1],
		ID:     id,
		PID:    b.pid,
	})
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("select pg_notify('%s', $1)", b.channel)
	_, err = b.pool.Exec(context.Background(), sql, msg)
	if err != nil {
		return err
	}
	return nil
}

// receive handles notifications from postgres
func (b *broker) receive(ctx context.Context, notification *pgconn.Notification) error {
	var msg message
	if err := json.Unmarshal([]byte(notification.Payload), &msg); err != nil {
		return err
	}

	// skip notifications that this process sent.
	if msg.PID == b.pid {
		return nil
	}

	getter, ok := b.tables[msg.Table]
	if !ok {
		return fmt.Errorf("unregistered table: %s", msg.Table)
	}
	payload, err := getter.GetByID(ctx, msg.ID)
	if err != nil {
		return fmt.Errorf("retrieving resource: %w", err)
	}

	b.localPublish(otf.Event{
		Type:    otf.EventType(fmt.Sprintf("%s_%s", msg.Table, msg.Action)),
		Payload: payload,
	})

	return nil
}
