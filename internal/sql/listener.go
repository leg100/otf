package sql

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"
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

		db         *DB                  // pool from which to acquire a dedicated connection to postgres
		mu         sync.Mutex           // sync access to maps
		forwarders map[string]TableFunc // maps table name to getter
	}

	// TableFunc is capable of converting a database row into a go type
	TableFunc func(action Action, record json.RawMessage)
)

func NewListener(logger logr.Logger, db *DB) *Listener {
	return &Listener{
		Logger:     logger.WithValues("component", "listener"),
		db:         db,
		forwarders: make(map[string]TableFunc),
	}
}

func (b *Listener) RegisterTable(table string, getter TableFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.forwarders[table] = getter
}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker. The islistening channel is closed once the broker has
// started listening; from this point onwards published messages will be
// forwarded.
func (b *Listener) Start(ctx context.Context) error {
	// Poll for new events every second.
	ticker := time.NewTicker(time.Second)
	for {
		rows := b.db.Query(ctx, "DELETE FROM events RETURNING *")
		events, err := pgx.CollectRows[event](rows, pgx.RowToStructByName)
		if err != nil {
			return err
		}
		for _, event := range events {
			forwarder, ok := b.forwarders[string(event.Table)]
			if !ok {
				b.Error(nil, "no getter found for table: %s", event.Table)
				continue
			}
			forwarder(event.Action, event.Payload)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return nil
		}
	}
}

// event is the insertion/update/deletion of a database row.
type event struct {
	Table   string          `json:"table"`   // pg table associated with change
	Action  Action          `json:"action"`  // INSERT/UPDATE/DELETE
	Payload json.RawMessage `json:"payload"` // the changed resource
}
