package repo

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql/pggen"
)

// PurgerLockID is a unique ID guaranteeing only one purger on a cluster is running at any time.
const PurgerLockID int64 = 179366396344335598

type (
	// Purge purges webhooks that are no longer in use.
	Purger struct {
		DB purgerDB

		logr.Logger
		pubsub.Subscriber
		Service
	}

	purgerDB interface {
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
		Tx(ctx context.Context, callback func(context.Context, pggen.Querier) error) error
	}

	repoConnectionEvent struct {
		webhookID uuid.UUID
	}
)

// Start starts the purger daemon. Should be invoked in a go routine.
func (p *Purger) Start(ctx context.Context) error {
	// Unsubscribe whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to webhook database events
	sub, err := p.Subscriber.Subscribe(ctx, "purger-")
	if err != nil {
		return err
	}

	// deleted unreferenced webhooks at startup in case:
	// (a) any unreferenced webhooks exist prior to the introduction of this
	// purger
	// (b) an error occured between the database sending an event and the purger
	// acting on the event (in which the case the purger should have restarted
	// and this will be re-run).
	if err := p.deleteUnreferencedWebhooks(ctx); err != nil {
		return err
	}

	for event := range sub {
		_, ok := event.Payload.(*repoConnectionEvent)
		if !ok {
			continue
		}
		if event.Type != pubsub.DeletedEvent {
			continue
		}

		// only repo connection deletion events reach this point

		if err := p.deleteUnreferencedWebhooks(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *Purger) deleteUnreferencedWebhooks(ctx context.Context) error {
	// Advisory lock ensures only one purger deletes the webhook from the cloud
	// provider.
	return p.DB.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		var locked bool
		err := p.DB.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock($1)", PurgerLockID).Scan(&locked)
		if err != nil {
			return err
		}
		if !locked {
			// Another purger obtained the lock first
			return nil
		}

		return p.Service.deleteUnreferencedWebhooks(ctx)
	})
}
