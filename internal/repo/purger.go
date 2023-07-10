package repo

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/vcsprovider"
)

// PurgerLockID is a unique ID guaranteeing only one purger on a cluster is running at any time.
const PurgerLockID int64 = 179366396344335598

type (
	// Purge purges webhooks that are no longer in use.
	Purger struct {
		DB purgerDB

		logr.Logger
		pubsub.Subscriber
		vcsprovider.VCSProviderService
		Service

		*cache
	}

	purgerDB interface {
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
		Tx(ctx context.Context, callback func(context.Context, pggen.Querier) error) error
	}
)

// Start starts the purger daemon. Should be invoked in a go routine.
func (p *Purger) Start(ctx context.Context) error {
	// Unsubscribe whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to webhook database events
	sub, err := p.Subscribe(ctx, "purger-")
	if err != nil {
		return err
	}

	// populate cache with existing hooks and vcs providers
	cache, err := newCache(ctx, cacheOptions{
		VCSProviderService: p.VCSProviderService,
		Subscriber:         p.Subscriber,
		Service:            p.Service,
	})
	if err != nil {
		return err
	}
	p.cache = cache

	for event := range sub {
		if err := p.handle(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (p *Purger) handle(ctx context.Context, event pubsub.Event) error {
	var deletedHook *hook

	switch payload := event.Payload.(type) {
	case *vcsprovider.VCSProvider:
		switch event.Type {
		case pubsub.CreatedEvent, pubsub.UpdatedEvent:
			p.cache.setProvider(payload)
		}
		return nil
	case *hook:
		switch event.Type {
		case pubsub.CreatedEvent, pubsub.UpdatedEvent:
			p.cache.setHook(payload)
			return nil
		case pubsub.DeletedEvent:
			deletedHook = payload
		}
	default:
		return nil
	}

	// only hook deletion events reach this point

	hook := p.getHook(deletedHook.id)
	if hook == nil {
		p.Error(nil, "webhook not found in cache", "webhook_id", deletedHook.id)
		return nil
	}
	provider := p.getProvider(hook.vcsProviderID)
	if provider == nil {
		p.Error(nil, "vcs provider not found in cache", "repo", hook.identifier, "vcs_provider_id", hook.vcsProviderID)
		return nil
	}
	// Advisory lock ensures only one purger deletes the webhook from the cloud
	// provider.
	err := p.DB.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		var locked bool
		err := p.DB.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock($1)", PurgerLockID).Scan(&locked)
		if err != nil {
			return err
		}
		if !locked {
			// Another purger obtained the lock first
			return nil
		}
		client, err := provider.NewClient(ctx)
		if err != nil {
			return err
		}
		err = client.DeleteWebhook(ctx, cloud.DeleteWebhookOptions{
			Repo: hook.identifier,
			ID:   *hook.cloudID,
		})
		if err != nil {
			p.Error(err, "deleting webhook", "repo", hook.identifier, "cloud", hook.cloud)
		} else {
			p.V(0).Info("deleted webhook", "repo", hook.identifier, "cloud", hook.cloud)
		}
		// Failure to delete the webhook from the cloud provider is not deemed a
		// fatal error.
		return nil
	})
	if err != nil {
		return err
	}
	// Regardless of success deleting the webhook from the cloud provider,
	// delete the hook and optionally, vcs provider, from the cache.
	p.delete(hook.id)

	return nil
}
