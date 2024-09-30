package sql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"log/slog"

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
		forwarders map[string]ForwardFunc // maps table name to getter
	}

	// ForwardFunc handles forwarding the id and action onto subscribers.
	ForwardFunc func(ctx context.Context, id string, action Action)

	// database connection pool
	pool interface {
		Acquire(ctx context.Context) (*pgxpool.Conn, error)
		Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
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

// RegisterFunc registers a function that is capable of converting database
// events for the given table into an OTF event.
func (b *Listener) RegisterFunc(table string, getter ForwardFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.forwarders[table] = getter
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
		var pge event
		if err := json.Unmarshal([]byte(notification.Payload), &pge); err != nil {
			b.Error(err, "unmarshaling postgres notification")
			continue
		}
		forwarder, ok := b.forwarders[string(pge.Table)]
		if !ok {
			b.Error(nil, "no getter found for table: %s", pge.Table)
			continue
		}
		forwarder(ctx, pge.ID, pge.Action)
	}
}

func (b *Listener) Started() <-chan struct{} {
	return b.islistening
}

// event is a postgres notification triggered by a database change.
type event struct {
	Table  string `json:"table"`  // pg table associated with change
	Action Action `json:"action"` // INSERT/UPDATE/DELETE
	ID     string `json:"id"`     // ID of changed row
}

func (v *event) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", v.ID),
		slog.String("action", string(v.Action)),
		slog.String("table", v.Table),
	}
	return slog.GroupValue(attrs...)
}
