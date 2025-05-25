package sql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

const (
	InsertAction = "INSERT"
	UpdateAction = "UPDATE"
	DeleteAction = "DELETE"
)

// ErrSubscriptionTerminated is for use by subscribers to indicate that their
// subscription has been terminated by the broker.
var ErrSubscriptionTerminated = errors.New("broker terminated the subscription")

type (
	// Action is the action that was carried out on a database table
	Action string

	// Listener listens for postgres events
	Listener struct {
		logr.Logger

		conn        *DB           // pool from which to acquire a dedicated connection to postgres
		islistening chan struct{} // semaphore for stating when listener is actually listening or not.

		// subscriptions: keyed by table name, and each table can have many
		// subscriptions
		subs map[string]map[chan Event]struct{}
		mu   sync.Mutex // sync access to maps
	}
)

func NewListener(logger logr.Logger, conn *DB) *Listener {
	return &Listener{
		Logger:      logger.WithValues("component", "listener"),
		conn:        conn,
		islistening: make(chan struct{}),
		subs:        make(map[string]map[chan Event]struct{}),
	}
}

func (b *Listener) Subscribe(ctx context.Context, table string) (<-chan Event, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := make(chan Event, 100)
	b.subs[table][sub] = struct{}{}

	// when the context is canceled remove the subscriber
	go func() {
		<-ctx.Done()
		b.unsubscribe(table, sub)
	}()

	return sub, func() { b.unsubscribe(table, sub) }
}

func (b *Listener) unsubscribe(table string, sub chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subs[table][sub]; !ok {
		// already unsubscribed
		return
	}
	close(sub)
	delete(b.subs[table], sub)
}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker. The listening channel is closed once the broker has
// started listening; from this point onwards published messages will be
// forwarded.
func (b *Listener) Start(ctx context.Context) error {
	// Obtain lock to prevent subscriptions *before* listening for events.

	conn, err := b.conn.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("unable to acquire postgres connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "LISTEN events"); err != nil {
		return err
	}
	b.V(2).Info("listening for events")

	// Release lock to now permit subscriptions
	b.islistening <- struct{}{}

	defer func() {
		// No longer listening
		<-b.islistening
		// Close all subscriptions.
		for table, subs := range b.subs {
			for sub := range subs {
				b.unsubscribe(table, sub)
			}
		}
	}()

	g, ctx := errgroup.WithContext(ctx)

	// cleanup old events
	g.Go(func() error {
		return b.cleanup(ctx)
	})

	// check for new events
	g.Go(func() error {
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
			row := b.conn.Query(ctx, `
SELECT *
FROM events
WHERE id = $1
`, notification.Payload)
			event, err := pgx.CollectOneRow(row, pgx.RowToStructByName[Event])
			if err != nil {
				return fmt.Errorf("retrieving events: %w", err)
			}
			if err := json.Unmarshal([]byte(event.Record), &event); err != nil {
				b.Error(err, "unmarshaling postgres event")
				continue
			}

			var fullSubscribers []chan Event

			b.mu.Lock()
			for sub := range b.subs[string(event.Table)] {
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
				b.Error(nil, "unsubscribing full subscriber", "sub", name, "queue_length", 100)
				b.unsubscribe(event.Table, name)
			}
		}
	})
	return g.Wait()
}

func (b *Listener) Started() <-chan struct{} {
	return b.islistening
}

func (b *Listener) cleanup(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	for {
		// delete events older than one minute
		_, err := b.conn.conn(ctx).Exec(ctx, `DELETE FROM events WHERE time < (current_timestamp - interval '1 minute')`)
		if err != nil {
			return fmt.Errorf("cleaning up old events: %w", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}

}

// Event is a postgres event triggered by a database change.
type Event struct {
	ID     int             `db:"id"`
	Table  string          `db:"_table"` // pg table associated with change
	Action Action          // INSERT/UPDATE/DELETE
	Record json.RawMessage // the changed row
	Time   time.Time       // time at which event occured
}

func (v Event) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("action", string(v.Action)),
		slog.String("table", v.Table),
		slog.Time("time", v.Time),
	}
	return slog.GroupValue(attrs...)
}
