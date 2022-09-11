package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
)

const EventsChannelName = "events"

// PubSub implements a distributed 'pub-sub' service for events, using postgres as a central broker
type PubSub struct {
	// postgres notification channel name
	channel string
	// pool from which to acquire a dedicated connection to postgres
	pool *pgxpool.Pool
	// acquired dedicated connection to postgres
	conn *pgxpool.Conn
	// local pub sub service to forward messages to
	local otf.PubSubService
	// db used for retrieving objects from the database
	db otf.DB
	logr.Logger
}

// message is the schema of the payload for use in the postgres notification channel.
type message struct {
	// Table is the postgres table on which the event occured
	Table string `json:"relation"`
	// Action is the type of change made to the relation
	Action string `json:"action"`
	// ID is the primary key of the changed row
	ID string `json:"id"`
	// PID of the otfd process that sent this event
	PID int `json:"pid"`
}

func NewPubSub(logger logr.Logger, pool *pgxpool.Pool) (*PubSub, error) {
	if pool == nil {
		return nil, fmt.Errorf("postgres pool is nil")
	}
	return &PubSub{
		channel: EventsChannelName,
		local:   inmem.NewPubSub(),
		pool:    pool,
		Logger:  logger.WithValues("component", "pubsub"),
	}, nil
}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker.
//
// TODO: start is responsible for acquiring connection and releasing it, but
// Publish() relies on that connection being present and may be called before
// start! Need to make this more rigorous.
func (ps *PubSub) Start(ctx context.Context) error {
	var err error
	ps.conn, err = ps.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer ps.conn.Release()

	if _, err := ps.conn.Exec(ctx, "listen "+ps.channel); err != nil {
		return err
	}

	// TODO: retry upon error with exp backoff
	for {
		notification, err := ps.conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}

		msg := message{}
		if err := json.Unmarshal([]byte(notification.Payload), &msg); err != nil {
			ps.Error(err, "unmarshaling postgres notification")
			continue
		}

		if msg.PID == os.Getpid() {
			// skip notifications that this process sent.
			continue
		}

		event, err := ps.reassemble(ctx, msg)
		if err != nil {
			ps.Error(err, "unmarshaling postgres notification")
			continue
		}

		ps.local.Publish(event)
	}
}

// Publish sends an event to subscribers, via postgres to subscribers on
// other machines, and via the local broker to subscribers within the same
// process.
func (ps *PubSub) Publish(event otf.Event) {
	ps.local.Publish(event)

	msg, err := marshal(event)
	if err != nil {
		ps.Error(err, "marshaling event into postgres notification payload")
		return
	}
	sql := fmt.Sprintf("select pg_notify('%s', $1)", ps.channel)
	_, err = ps.pool.Exec(context.Background(), sql, msg)
	if err != nil {
		ps.Error(err, "sending postgres notification")
		return
	}
}

// Subscribe subscribes the caller to a stream of events.
func (ps *PubSub) Subscribe(ctx context.Context) <-chan otf.Event {
	return ps.local.Subscribe(ctx)
}

// reassemble a message into an otf event
//
// TODO: return pointer to event to indicate there is no event to public but no
// error occured (?)
func (ps *PubSub) reassemble(ctx context.Context, msg message) (otf.Event, error) {
	var payload any
	var err error
	switch msg.Table {
	case "run":
		payload, err = ps.db.GetRun(ctx, msg.ID)
		if err != nil {
			return otf.Event{}, err
		}
	case "workspace":
		payload, err = ps.db.GetWorkspace(ctx, otf.WorkspaceSpec{ID: &msg.ID})
		if err != nil {
			return otf.Event{}, err
		}
	default:
		// TODO: log error message
		return otf.Event{}, nil
	}
	return otf.Event{
		Type:    otf.EventType(fmt.Sprintf("%s_%s", msg.Table, msg.Action)),
		Payload: payload,
	}, nil
}

// marshal an otf event into a JSON-encoded message
func marshal(event otf.Event) ([]byte, error) {
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
		PID:    os.Getpid(),
	})
}
