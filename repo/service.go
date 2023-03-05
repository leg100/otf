package repo

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/sql"
)

type Service struct {
	logr.Logger
	otf.VCSProviderService

	db      *pgdb // access to hook database
	factory       // produce new hooks
}

func NewService(opts NewServiceOptions) *Service {
	factory := newFactory(opts.HostnameService, opts.CloudService)
	return &Service{
		Logger:             opts.Logger,
		VCSProviderService: opts.VCSProviderService,
		db:                 newPGDB(opts.DB, factory),
		factory:            factory,
	}
}

type NewServiceOptions struct {
	CloudService cloud.Service

	*sql.DB
	logr.Logger
	otf.HostnameService
	otf.VCSProviderService
}

// Connect an OTF resource to a VCS repo.
func (s *Service) Connect(ctx context.Context, opts otf.ConnectionOptions) error {
	unsynced, err := s.newHook(newHookOpts{
		identifier: opts.Identifier,
		cloud:      opts.Cloud,
	})
	if err != nil {
		return err
	}

	client, err := s.GetVCSClient(ctx, opts.VCSProviderID)
	if err != nil {
		return err
	}

	// lock webhooks table to prevent concurrent updates (a row-level lock is
	// insufficient)
	return s.db.lock(ctx, func(tx *pgdb) error {
		hook, err := tx.getOrCreate(ctx, unsynced)
		if err != nil {
			return err
		}

		if err := hook.sync(ctx, client); err != nil {
			return err
		}

		if err := tx.updateCloudID(ctx, hook.id, *hook.cloudID); err != nil {
			return err
		}

		return tx.createConnection(ctx, hook.id, opts)
	})
}

// NOTE: if the webhook cannot be deleted from the repo then this is not deemed
// fatal and the hook is still deleted from the database.
func (s *Service) Disconnect(ctx context.Context, opts otf.ConnectionOptions) error {
	// separately capture any error resulting from attempting to delete the
	// webhook from the VCS repo
	var repoErr error

	// lock webhooks table to prevent concurrent updates (a row-level lock is
	// insufficient)
	err := s.db.lock(ctx, func(tx *pgdb) error {
		// disconnect connected resource
		hookID, err := tx.deleteConnection(ctx, opts)
		if err != nil {
			return err
		}

		conns, err := tx.countConnections(ctx, hookID)
		if err != nil {
			return err
		}
		if *conns > 0 {
			return nil
		}

		// no more connections; delete webhook from both db and VCS provider

		hook, err := tx.deleteHook(ctx, hookID)
		if err != nil {
			return err
		}
		client, err := s.GetVCSClient(ctx, opts.VCSProviderID)
		if err != nil {
			return err
		}
		err = client.DeleteWebhook(ctx, cloud.DeleteWebhookOptions{
			Identifier: hook.identifier,
			ID:         *hook.cloudID,
		})
		if err != nil {
			repoErr = fmt.Errorf("%w: unable to delete webhook from repo: %w", otf.ErrWarning, err)
		} else {
			s.V(0).Info("deleted webhook", "repo", hook.identifier, "cloud", hook.cloud)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return repoErr
}
