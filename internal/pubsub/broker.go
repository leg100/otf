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
	"golang.org/x/exp/slog"
)

const (
	defaultChannel = "events"

	// subBufferSize is the buffer size of the channel for each subscription.
	subBufferSize = 100
)

type (
	// Getter retrieves an event payload using its ID.
	Getter interface {
		GetByID(context.Context, string, DBAction) (any, error)
	}

	// Broker is a pubsub Broker implemented using postgres' listen/notify
	Broker struct {
		logr.Logger

		channel     string        // postgres notification channel name
		pool        pool          // pool from which to acquire a dedicated connection to postgres
		islistening chan struct{} // semaphore that's closed once broker is listening

		subs    map[string]chan Event       // subscriptions
		metrics map[string]prometheus.Gauge // metric for each subscription
		mu      sync.Mutex                  // sync access to maps

		*converter
	}

	pool interface {
		Acquire(ctx context.Context) (*pgxpool.Conn, error)
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	}

	// pgevent is the payload of a postgres notification triggered by a database
	// change.
	pgevent struct {
		Table  string   `json:"table"`  // pg table associated with change
		Action DBAction `json:"action"` // INSERT/UPDATE/DELETE
		ID     string   `json:"id"`     // id of changed row
	}
)

func NewBroker(logger logr.Logger, db pool) *Broker {
	return &Broker{
		Logger:      logger.WithValues("component", "broker"),
		pool:        db,
		islistening: make(chan struct{}),
		channel:     defaultChannel,
		subs:        make(map[string]chan Event),
		metrics:     make(map[string]prometheus.Gauge),
		converter:   newConverter(),
	}
}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker. The listening channel is closed once the broker has
// started listening; from this point onwards published messages will be
// forwarded.
func (b *Broker) Start(ctx context.Context) error {
	conn, err := b.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("unable to acquire postgres connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "listen "+b.channel); err != nil {
		return err
	}
	b.V(2).Info("listening for events")
	close(b.islistening) // close semaphore to indicate broker is now listening

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
		var pge pgevent
		if err := json.Unmarshal([]byte(notification.Payload), &pge); err != nil {
			b.Error(err, "unmarshaling postgres notification")
			continue
		}
		event, err := b.convert(ctx, pge)
		if err != nil {
			b.Error(err, "converting postgres notification into event", "notification", pge)
			continue
		}
		b.localPublish(event)
	}
}

func (b *Broker) Started() <-chan struct{} {
	return b.islistening
}

// Publish sends an event to subscribers.
func (b *Broker) Publish(event Event) {
	// ignore non-local publishing of events; the database itself is now
	// responsible for triggering the publishing of events
	if !event.Local {
		return
	}
	// send event only to local subscribers
	b.localPublish(event)
}

// Subscribe subscribes the caller to a stream of events. Prefix is an
// identifier prefixed to a random string to helpfully identify the subscriber
// in metrics.
func (b *Broker) Subscribe(ctx context.Context, prefix string) (<-chan Event, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	name := prefix + internal.GenerateRandomString(4)

	sub := make(chan Event, subBufferSize)
	if _, ok := b.subs[name]; ok {
		return nil, fmt.Errorf("name already taken")
	}
	b.subs[name] = sub

	totalSubscribers.Inc()

	b.metrics[name] = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "otf",
		Subsystem:   "pub_sub",
		Name:        "queue_length",
		Help:        "Number of items in subscriber's queue",
		ConstLabels: prometheus.Labels{"name": name},
	})
	if err := prometheus.Register(b.metrics[name]); err != nil {
		return nil, fmt.Errorf("registering metric for subscriber: %s: %w", name, err)
	}

	// when the context is canceled remove the subscriber
	go func() {
		<-ctx.Done()
		b.unsubscribe(name)
	}()

	return sub, nil
}

func (b *Broker) unsubscribe(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub, ok := b.subs[name]
	if !ok {
		// already unsubscribed
		return
	}

	totalSubscribers.Dec()

	close(sub)
	delete(b.subs, name)

	prometheus.Unregister(b.metrics[name])
	delete(b.metrics, name)
}

// localPublish publishes an event to subscribers on the local node
func (b *Broker) localPublish(event Event) {
	var fullSubscribers []string

	b.mu.Lock()
	for name, sub := range b.subs {
		// record sub's chan size
		b.metrics[name].Set(float64(len(sub)))

		select {
		case sub <- event:
			continue
		default:
			// could not publish event to subscriber because their buffer is
			// full, so add them to a list for action below
			fullSubscribers = append(fullSubscribers, name)
		}
	}
	b.mu.Unlock()

	// forceably unsubscribe full subscribers and let the client re-subscribe.
	for _, name := range fullSubscribers {
		b.Error(nil, "unsubscribing full subscriber", "sub", name, "queue_length", subBufferSize)
		b.unsubscribe(name)
	}
}

func (v *pgevent) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", v.ID),
		slog.String("action", string(v.Action)),
		slog.String("table", v.Table),
	}
	return slog.GroupValue(attrs...)
}
