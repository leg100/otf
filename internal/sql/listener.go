package sql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/logr"
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
		islistening chan struct{} // semaphore that's populated once listener is listening.

		mu   sync.Mutex             // sync access to maps
		subs map[Table]chan<- Event // table event subscriptions
	}

	// ForwardFunc forwards an event to a client
	ForwardFunc func(event Event)
)

func NewListener(logger logr.Logger, conn *DB) *Listener {
	return &Listener{
		Logger:      logger.WithValues("component", "listener"),
		conn:        conn,
		islistening: make(chan struct{}, 1),
		subs:        make(map[Table]chan<- Event),
	}
}

// Start the database event listener.
func (b *Listener) Start(ctx context.Context) error {
	// Close listen connection when leaving func.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Unsubscribe subscribers upon exit.
	defer func() {
		b.Logger.Info("unsubscribing subscribers")
		b.mu.Lock()
		for k, sub := range b.subs {
			close(sub)
			delete(b.subs, k)
		}
		b.mu.Unlock()
		b.Logger.Info("finished unsubscribing subscribers")
	}()

	notifications, err := b.conn.Listen(ctx, "events")
	if err != nil {
		return fmt.Errorf("listening to postgres notification channel: %w", err)
	}

	// Inform caller that we're now listening. This routine may be called
	// more than once if the listener is restarted, e.g. there is a
	// transient database failure. Therefore we don't block on this channel
	// if a message has already been published by a previous start.
	select {
	case b.islistening <- struct{}{}:
	default:
		b.Logger.Info("now listening")
	}

	b.V(2).Info("listening for events")

	g, ctx := errgroup.WithContext(ctx)

	// cleanup old events
	g.Go(func() error {
		return b.cleanup(ctx)
	})

	// check for new events
	g.Go(func() error {
		for notification := range notifications {
			row := b.conn.Query(ctx, `SELECT * FROM events WHERE id = $1 `, notification)
			event, err := pgx.CollectOneRow(row, pgx.RowToStructByName[Event])
			if err != nil {
				return fmt.Errorf("retrieving events: %w", err)
			}
			// What is this doing??
			if err := json.Unmarshal([]byte(event.Record), &event); err != nil {
				b.Error(err, "unmarshaling postgres event")
				continue
			}
			// 1. use table to lookup decoder in a mapping of table to decoder
			// 2. decode message using decoder
			b.mu.Lock()
			forwarder, ok := b.subs[event.Table]
			b.mu.Unlock()
			if !ok {
				b.Error(nil, "no event forwarder found", "table", event.Table)
				continue
			}
			forwarder <- event
		}
		return nil
	})
	return g.Wait()
}

func (b *Listener) Started() <-chan struct{} {
	return b.islistening
}

// Subscribe to events for a database table.
func (b *Listener) Subscribe(table Table) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := make(chan Event)
	b.subs[table] = sub
	return sub
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
	Table  Table           `db:"_table"` // pg table associated with change
	Action Action          // INSERT/UPDATE/DELETE
	Record json.RawMessage // the changed row
	Time   time.Time       // time at which event occured
}

func (v Event) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("action", string(v.Action)),
		slog.Any("table", v.Table),
		slog.Time("time", v.Time),
	}
	return slog.GroupValue(attrs...)
}
