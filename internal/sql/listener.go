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
		islistening chan struct{} // semaphore that's closed once broker is listening

		mu         sync.Mutex             // sync access to maps
		forwarders map[string]ForwardFunc // maps table name to event forwarder
	}

	// ForwardFunc forwards an event to a client
	ForwardFunc func(event Event)
)

func NewListener(logger logr.Logger, conn *DB) *Listener {
	return &Listener{
		Logger:      logger.WithValues("component", "listener"),
		conn:        conn,
		islistening: make(chan struct{}),
		forwarders:  make(map[string]ForwardFunc),
	}
}

// RegisterTable maps a table to an event forwarding function: the function
// henceforth is called with events triggered on the table.
func (b *Listener) RegisterTable(table string, forwarder ForwardFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.forwarders[table] = forwarder
}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker. The listening channel is closed once the broker has
// started listening; from this point onwards published messages will be
// forwarded.
func (b *Listener) Start(ctx context.Context) error {
	// Cancel listen connection when leaving func.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	notifications, err := b.conn.Listen(ctx, "events")
	if err != nil {
		return fmt.Errorf("listening to postgres notification channel: %w", err)
	}

	b.V(2).Info("listening for events")
	go func() {
		// close semaphore to indicate broker is now listening
		b.islistening <- struct{}{}
	}()

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
			if err := json.Unmarshal([]byte(event.Record), &event); err != nil {
				b.Error(err, "unmarshaling postgres event")
				continue
			}
			forwarder, ok := b.forwarders[string(event.Table)]
			if !ok {
				b.Error(nil, "no event forwarder found", "table", event.Table)
				continue
			}
			forwarder(event)
		}
		return nil
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
