package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"gopkg.in/cenkalti/backoff.v1"
)

const defaultChannel = "events"

// a unique identity string for distinguishing this process from other otfd
// processes
var pid = uuid.New().String()

// Getter retrieves a resource using its ID.
type Getter interface {
	GetByID(context.Context, string) (any, error)
}

// Hub is a pubsub Hub implemented using postgres' listen/notify
type Hub struct {
	logr.Logger

	channel string            // postgres notification channel name
	pool    *pgxpool.Pool     // pool from which to acquire a dedicated connection to postgres
	local   otf.PubSubService // local pub sub service to forward messages to
	pid     string            // each pubsub maintains a unique identifier to distriguish messages it
	// sends from messages other pubsub have sent
	tables map[string]Getter // registered means of reassembling back into events
}

type HubConfig struct {
	ChannelName *string
	PID         *string
	PoolDB      otf.DB
}

func NewHub(logger logr.Logger, cfg HubConfig) (*Hub, error) {
	// required config
	if cfg.PoolDB == nil {
		return nil, errors.New("missing database connection pool")
	}

	ps := &Hub{
		Logger:  logger.WithValues("component", "pubsub"),
		pid:     pid,
		channel: defaultChannel,
		local:   newSpoke(),
		tables:  make(map[string]Getter),
	}

	pool, err := cfg.PoolDB.Pool()
	if err != nil {
		return nil, err
	}
	ps.pool = pool

	// optional config
	if cfg.ChannelName != nil {
		ps.channel = *cfg.ChannelName
	}
	if cfg.PID != nil {
		ps.pid = *cfg.PID
	}

	return ps, nil
}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker.
func (b *Hub) Start(ctx context.Context) error {
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

			msg := message{}
			if err := json.Unmarshal([]byte(notification.Payload), &msg); err != nil {
				b.Error(err, "unmarshaling postgres notification")
				continue
			}

			if msg.PID == b.pid {
				// skip notifications that this process sent.
				continue
			}

			event, err := b.reassemble(ctx, msg)
			if err != nil {
				b.Error(err, "unmarshaling postgres notification")
				continue
			}

			b.local.Publish(event)
		}
	}
	return backoff.RetryNotify(op, backoff.NewExponentialBackOff(), nil)
}

// Publish sends an event to subscribers, via postgres to subscribers on
// other machines, and via the local broker to subscribers within the same
// process.
func (b *Hub) Publish(event otf.Event) {
	b.local.Publish(event)

	// Don't publish vcs events to the rest of the cluster: the triggerer
	// subscribes to vcs events and runs on each node in the cluster, but we
	// only want each event to trigger one triggerer, so we restrict publishing
	// the event to the local node.
	//
	// TODO: this is naff, but we'll remove this once we refactor out the
	// triggerer with something better.
	if event.Type == otf.EventVCS {
		return
	}

	msg, err := b.marshal(event)
	if err != nil {
		b.Error(err, "marshaling event into postgres notification payload")
		return
	}
	sql := fmt.Sprintf("select pg_notify('%s', $1)", b.channel)
	_, err = b.pool.Exec(context.Background(), sql, msg)
	if err != nil {
		b.Error(err, "sending postgres notification")
		return
	}
}

// Subscribe subscribes the caller to a stream of events.
func (b *Hub) Subscribe(ctx context.Context, name string) (<-chan otf.Event, error) {
	return b.local.Subscribe(ctx, name)
}

// Register a means of reassembling a postgres message back into an otf event
func (b *Hub) Register(table string, getter Getter) {
	b.tables[table] = getter
}

// reassemble a postgres message into an otf event
func (b *Hub) reassemble(ctx context.Context, msg message) (otf.Event, error) {
	getter, ok := b.tables[msg.Table]
	if !ok {
		return otf.Event{}, fmt.Errorf("unregistered table: %s", msg.Table)
	}
	payload, err := getter.GetByID(ctx, msg.ID)
	if err != nil {
		return otf.Event{}, err
	}
	return otf.Event{
		Type:    otf.EventType(fmt.Sprintf("%s_%s", msg.Table, msg.Action)),
		Payload: payload,
	}, nil
}

// marshal an otf event into a JSON-encoded postgres message
func (b *Hub) marshal(event otf.Event) ([]byte, error) {
	id, hasID := otf.GetID(event.Payload)
	if !hasID {
		return nil, fmt.Errorf("cannot marshal event without an identifiable payload")
	}
	parts := strings.SplitN(string(event.Type), "_", 2)
	if len(parts) < 2 {
		// log message
		return nil, fmt.Errorf("event has an invalid type format: %s", event.Type)
	}
	return json.Marshal(&message{
		Table:  parts[0],
		Action: parts[1],
		ID:     id,
		PID:    b.pid,
	})
}
