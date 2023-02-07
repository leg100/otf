package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"gopkg.in/cenkalti/backoff.v1"
)

const defaultChannel = "events"

// a unique identity string for distinguishing this process from other otfd
// processes
var pid = uuid.New().String()

// hub is a pubsub hub implemented using postgres' listen/notify
type hub struct {
	logr.Logger
	DB // db used for retrieving event payloads

	channel string            // postgres notification channel name
	pool    *pgxpool.Pool     // pool from which to acquire a dedicated connection to postgres
	local   otf.PubSubService // local pub sub service to forward messages to
	pid     string            // each pubsub maintains a unique identifier to distriguish messages it
	// sends from messages other pubsub have sent
}

type HubConfig struct {
	ChannelName *string
	PID         *string
	PoolDB      otf.DB
	DB          DB
}

type (
	DB interface {
		GetOrganizationByID(context.Context, string) (otf.Organization, error)
		GetRun(context.Context, string) (otf.Run, error)
		GetWorkspace(context.Context, string) (otf.Workspace, error)
		GetChunkByID(context.Context, int) (otf.Chunk, error)
	}
)

func NewHub(logger logr.Logger, cfg HubConfig) (*hub, error) {
	// required config
	if cfg.PoolDB == nil {
		return nil, errors.New("missing database connection pool")
	}
	if cfg.DB == nil {
		return nil, errors.New("missing database")
	}

	ps := &hub{
		Logger:  logger.WithValues("component", "pubsub"),
		pid:     pid,
		channel: defaultChannel,
		local:   inmem.NewPubSub(),
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

// message is the schema of the payload for use in the postgres notification channel.
type message struct {
	// Table is the postgres table on which the event occured
	Table string `json:"relation"`
	// Action is the type of change made to the relation
	Action string `json:"action"`
	// ID is the primary key of the changed row
	ID string `json:"id"`
	// PID is the process id that sent this event
	PID string `json:"pid"`
}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker.
func (b *hub) Start(ctx context.Context) error {
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
func (b *hub) Publish(event otf.Event) {
	b.local.Publish(event)

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
func (b *hub) Subscribe(ctx context.Context, name string) (<-chan otf.Event, error) {
	return b.local.Subscribe(ctx, name)
}

// reassemble a postgres message into an otf event
func (b *hub) reassemble(ctx context.Context, msg message) (otf.Event, error) {
	var payload any
	var err error

	switch msg.Table {
	case "organization":
		payload, err = b.GetOrganizationByID(ctx, msg.ID)
		if err != nil {
			return otf.Event{}, err
		}
	case "run":
		payload, err = b.GetRun(ctx, msg.ID)
		if err != nil {
			return otf.Event{}, err
		}
	case "workspace":
		payload, err = b.GetWorkspace(ctx, msg.ID)
		if err != nil {
			return otf.Event{}, err
		}
	case "log":
		id, err := strconv.Atoi(msg.ID)
		if err != nil {
			return otf.Event{}, fmt.Errorf("converting chunk ID from string to an integer: %w", err)
		}
		payload, err = b.GetChunkByID(ctx, id)
		if err != nil {
			return otf.Event{}, err
		}
	default:
		return otf.Event{}, fmt.Errorf("unknown table specified in events notification: %s", msg.Table)
	}
	return otf.Event{
		Type:    otf.EventType(fmt.Sprintf("%s_%s", msg.Table, msg.Action)),
		Payload: payload,
	}, nil
}

// marshal an otf event into a JSON-encoded postgres message
func (b *hub) marshal(event otf.Event) ([]byte, error) {
	obj, ok := event.Payload.(otf.Identity)
	if !ok {
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
		ID:     obj.ID(),
		PID:    b.pid,
	})
}
