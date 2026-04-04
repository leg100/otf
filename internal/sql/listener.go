package sql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		logger      logr.Logger
		conn        *DB           // pool from which to acquire a dedicated connection to postgres
		islistening chan struct{} // semaphore that's populated once listener is listening.
		brokers     map[Table]Broker
	}

	// ForwardFunc forwards an event to a client
	ForwardFunc func(event Event)

	// Broker is a broker handling events for a specific table.
	Broker interface {
		// Forward database event to broker
		Forward(Event)
		// Table retrieves the table the broker is handling events for.
		Table() Table
		// Enable the broker.
		Enable()
		// Disable the broker.
		Disable()
	}
)

func NewListener(logger logr.Logger, conn *DB, brokers ...Broker) *Listener {
	listener := &Listener{
		logger:      logger.WithValues("component", "listener"),
		conn:        conn,
		islistening: make(chan struct{}, 1),
	}
	listener.brokers = make(map[Table]Broker, len(brokers))
	for _, broker := range brokers {
		listener.brokers[broker.Table()] = broker
	}
	return listener

}

// Start the pubsub daemon; listen to notifications from postgres and forward to
// local pubsub broker. The listening channel is closed once the broker has
// started listening; from this point onwards published messages will be
// forwarded.
func (l *Listener) Start(ctx context.Context) error {
	// Close listen connection when leaving func.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	notifications, err := l.conn.Listen(ctx, "events")
	if err != nil {
		return fmt.Errorf("listening to postgres notification channel: %w", err)
	}

	// Inform caller that we're now listening. This routine may be called
	// more than once if the listener is restarted, e.g. there is a
	// transient database failure. Therefore we don't block on this channel
	// if a message has already been published by a previous start.
	select {
	case l.islistening <- struct{}{}:
	default:
		l.logger.Info("now listening")
	}

	// Now that we're listening inform the brokers.
	for _, broker := range l.brokers {
		broker.Enable()
	}

	g, ctx := errgroup.WithContext(ctx)

	// cleanup old events
	g.Go(func() error {
		return l.cleanup(ctx)
	})

	// check for new events
	g.Go(func() error {
		for notification := range notifications {
			row := l.conn.Query(ctx, `SELECT * FROM events WHERE id = $1 `, notification)
			event, err := pgx.CollectOneRow(row, pgx.RowToStructByName[Event])
			if err != nil {
				return fmt.Errorf("retrieving events: %w", err)
			}
			// What is this doing??
			if err := json.Unmarshal([]byte(event.Record), &event); err != nil {
				l.logger.Error(err, "unmarshaling postgres event")
				continue
			}
			broker, ok := l.brokers[event.Table]
			if !ok {
				l.logger.Error(nil, "no event broker found for table", "table", event.Table)
				continue
			}
			broker.Forward(event)
		}
		// Connection to notification channel closed; inform the brokers so that
		// they can inform their subscribers).
		l.logger.Info("disabling brokers")
		for _, broker := range l.brokers {
			broker.Disable()
		}
		l.logger.Info("exiting listener")
		return fmt.Errorf("connection to database notification channel terminated")
	})
	return g.Wait()
}

func (l *Listener) Started() <-chan struct{} {
	return l.islistening
}

func (l *Listener) cleanup(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	for {
		// delete events older than one minute
		_, err := l.conn.conn(ctx).Exec(ctx, `DELETE FROM events WHERE time < (current_timestamp - interval '1 minute')`)
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
