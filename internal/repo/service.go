package repo

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/vcsprovider"
)

type (
	RepoService = Service

	// Service manages VCS repositories
	Service interface {
		// Connect adds a connection between a VCS repo and an OTF resource. A
		// webhook is created if one doesn't exist already.
		Connect(ctx context.Context, opts ConnectOptions) (*Connection, error)
		// Disconnect removes a connection between a VCS repo and an OTF
		// resource. If there are no more connections then its
		// webhook is removed.
		Disconnect(ctx context.Context, opts DisconnectOptions) error
	}

	service struct {
		logr.Logger
		vcsprovider.Service
		db *pgdb

		*handler      // handles incoming vcs events
		factory       // produce new hooks
		*synchroniser // synchronise hooks
	}

	Options struct {
		logr.Logger

		CloudService cloud.Service

		internal.DB
		internal.HostnameService
		pubsub.Publisher
		VCSProviderService vcsprovider.Service
	}
)

func NewService(opts Options) *service {
	factory := newFactory(opts.HostnameService, opts.CloudService)
	db := newPGDB(opts.DB, factory)
	handler := &handler{
		Logger:    opts.Logger,
		Publisher: opts.Publisher,
		db:        db,
	}
	return &service{
		Logger:       opts.Logger,
		Service:      opts.VCSProviderService,
		db:           db,
		factory:      factory,
		handler:      handler,
		synchroniser: &synchroniser{Logger: opts.Logger},
	}
}

// Connect an OTF resource to a VCS repo.
func (s *service) Connect(ctx context.Context, opts ConnectOptions) (*Connection, error) {
	vcsProvider, err := s.GetVCSProvider(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, fmt.Errorf("retrieving vcs provider: %w", err)
	}
	client, err := s.GetVCSClient(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, fmt.Errorf("retrieving vcs client: %w", err)
	}
	_, err = client.GetRepository(ctx, opts.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("checking repository exists: %w", err)
	}

	hook, err := s.newHook(newHookOpts{
		identifier: opts.RepoPath,
		cloud:      vcsProvider.CloudConfig.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("constructing webhook: %w", err)
	}

	// allow caller to provide their own tx
	var db *pgdb
	if opts.Tx != nil {
		db = newPGDB(opts.Tx, s.factory)
	} else {
		db = s.db
	}

	// lock webhooks table to prevent concurrent updates (a row-level lock is
	// insufficient)
	err = db.lock(ctx, func(tx *pgdb) error {
		hook, err = tx.getOrCreateHook(ctx, hook)
		if err != nil {
			return fmt.Errorf("getting or creating webhook: %w", err)
		}
		if err := s.sync(ctx, tx, client, hook); err != nil {
			return fmt.Errorf("synchronising webhook: %w", err)
		}
		return tx.createConnection(ctx, hook.id, opts)
	})
	if err != nil {
		return nil, err
	}
	return &Connection{
		Repo:          opts.RepoPath,
		VCSProviderID: opts.VCSProviderID,
	}, nil
}

// Disconnect resource from repo
//
// NOTE: if the webhook cannot be deleted from the repo then this is not deemed
// fatal and the hook is still deleted from the database.
func (s *service) Disconnect(ctx context.Context, opts DisconnectOptions) error {
	// allow caller to provide their own tx
	var db *pgdb
	if opts.Tx != nil {
		db = newPGDB(opts.Tx, s.factory)
	} else {
		db = s.db
	}

	// lock webhooks table to prevent concurrent updates (a row-level lock is
	// insufficient)
	return db.lock(ctx, func(tx *pgdb) error {
		hookID, vcsProviderID, err := tx.deleteConnection(ctx, opts)
		if err != nil {
			return err
		}

		conns, err := tx.countConnections(ctx, hookID)
		if err != nil {
			return err
		}
		if conns > 0 {
			return nil
		}

		// no more connections; delete webhook from both db and VCS provider

		hook, err := tx.deleteHook(ctx, hookID)
		if err != nil {
			return err
		}
		client, err := s.GetVCSClient(ctx, vcsProviderID)
		if err != nil {
			return err
		}
		err = client.DeleteWebhook(ctx, cloud.DeleteWebhookOptions{
			Repo: hook.identifier,
			ID:   *hook.cloudID,
		})
		if err != nil {
			// failure to delete webhook from cloud is not fatal
			s.Error(err, "deleting webhook", "repo", hook.identifier, "cloud", hook.cloud)
		} else {
			s.V(0).Info("deleted webhook", "repo", hook.identifier, "cloud", hook.cloud)
		}
		return nil
	})
}
