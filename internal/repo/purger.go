package repo

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/vcsprovider"
)

// PurgerLockID is a unique ID guaranteeing only one purger on a cluster is running at any time.
const PurgerLockID int64 = 179366396344335598

type (
	// Purge purges webhooks that are no longer in use.
	Purger struct {
		logr.Logger
		pubsub.Subscriber
		vcsprovider.VCSProviderService
		*sql.DB
	}
)

// Start starts the purger daemon. Should be invoked in a go routine.
//
// NOTE: if the webhook cannot be deleted from the repo then this is not deemed
// fatal and the hook is still deleted from the database.
func (r *Purger) Start(ctx context.Context) error {
	// Unsubscribe whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to webhook database events
	sub, err := r.Subscribe(ctx, "purger-")
	if err != nil {
		return err
	}
	for event := range sub {
		hook, ok := event.Payload.(*hook)
		if !ok {
			// Skip non-hook events
			continue
		}
		if event.Type != pubsub.DeletedEvent {
			// Skip non-deleted events
			continue
		}
		// In case there are multiple purgers running the following advisory
		// lock ensures only one purger is successfully granted the lock and
		// therefore only one purger deletes the webhook from the cloud provider.
		err = r.DB.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
			_, err := r.DB.Exec(ctx, "SELECT pg_try_advisory_xact_lock($1)", PurgerLockID)
			if err != nil {
				return err
			}
			client, err := r.GetVCSClient(ctx, hook.vcsProviderID)
			if err != nil {
				return err
			}
			err = client.DeleteWebhook(ctx, cloud.DeleteWebhookOptions{
				Repo: hook.identifier,
				ID:   *hook.cloudID,
			})
			if err != nil {
				r.Error(err, "deleting webhook", "repo", hook.identifier, "cloud", hook.cloud)
			} else {
				r.V(0).Info("deleted webhook", "repo", hook.identifier, "cloud", hook.cloud)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
