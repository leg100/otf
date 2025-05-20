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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultChannel = "events"

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

		channel     string        // postgres notification channel name
		pool        pool          // pool from which to acquire a dedicated connection to postgres
		islistening chan struct{} // semaphore that's closed once broker is listening

		mu         sync.Mutex             // sync access to maps
		forwarders map[string]ForwardFunc // maps table name to event forwarder
	}

	// ForwardFunc forwards an event to a client
	ForwardFunc func(event Event)

	// database connection pool
	pool interface {
		Acquire(ctx context.Context) (*pgxpool.Conn, error)
		Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	}
)

func NewListener(logger logr.Logger, db pool) *Listener {
	return &Listener{
		Logger:      logger.WithValues("component", "listener"),
		pool:        db,
		islistening: make(chan struct{}),
		channel:     defaultChannel,
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
		var event Event
		if err := json.Unmarshal([]byte(notification.Payload), &event); err != nil {
			b.Error(err, "unmarshaling postgres notification")
			continue
		}
		forwarder, ok := b.forwarders[string(event.Table)]
		if !ok {
			b.Error(nil, "no getter found for table: %s", event.Table)
			continue
		}
		forwarder(event)
	}
}

func (b *Listener) Started() <-chan struct{} {
	return b.islistening
}

// Event is a postgres event triggered by a database change.
type Event struct {
	Table  string          `json:"table"`  // pg table associated with change
	Action Action          `json:"action"` // INSERT/UPDATE/DELETE
	Record json.RawMessage `json:"record"` // the changed row
	Time   time.Time       `json:"time"`   // time at which event occured
}

func (v *Event) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("action", string(v.Action)),
		slog.String("table", v.Table),
		slog.Time("time", v.Time),
	}
	return slog.GroupValue(attrs...)
}
