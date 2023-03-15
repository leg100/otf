package repo

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/vcsprovider"
)

type (
	Service struct {
		logr.Logger
		vcsprovider.Service
		otf.DB

		factory // produce new hooks
		*handler
	}

	Options struct {
		CloudService cloud.Service

		otf.DB
		logr.Logger
		otf.HostnameService
		otf.Publisher
		VCSProviderService vcsprovider.Service
	}
)

func NewService(opts Options) *Service {
	factory := newFactory(opts.HostnameService, opts.CloudService)
	handler := &handler{
		Logger:    opts.Logger,
		Publisher: opts.Publisher,
		db:        newPGDB(opts.DB, factory),
	}
	return &Service{
		Logger:  opts.Logger,
		Service: opts.VCSProviderService,
		DB:      opts.DB,
		factory: factory,
		handler: handler,
	}
}

// Connect an OTF resource to a VCS repo.
func (s *Service) Connect(ctx context.Context, opts otf.ConnectOptions) (*otf.Connection, error) {
	vcsProvider, err := s.GetVCSProvider(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, err
	}
	client, err := s.GetVCSClient(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// allow caller to provide their own tx
	var db *pgdb
	if opts.Tx != nil {
		db = newPGDB(opts.Tx, s.factory)
	} else {
		db = newPGDB(s.DB, s.factory)
	}

	// lock webhooks table to prevent concurrent updates (a row-level lock is
	// insufficient)
	err = db.lock(ctx, func(tx *pgdb) error {
		hook, err = tx.getOrCreateHook(ctx, hook)
		if err != nil {
			return err
		}

		if err := hook.sync(ctx, client); err != nil {
			return err
		}

		if err := tx.updateHookCloudID(ctx, hook.id, *hook.cloudID); err != nil {
			return err
		}

		return tx.createConnection(ctx, hook.id, opts)
	})
	return &otf.Connection{
		Repo:          opts.RepoPath,
		VCSProviderID: opts.VCSProviderID,
	}, nil
}

// Disconnect resource from repo
//
// NOTE: if the webhook cannot be deleted from the repo then this is not deemed
// fatal and the hook is still deleted from the database.
func (s *Service) Disconnect(ctx context.Context, opts otf.DisconnectOptions) error {
	// separately capture any error resulting from attempting to delete the
	// webhook from the VCS repo
	var repoErr error

	// allow caller to provide their own tx
	var db *pgdb
	if opts.Tx != nil {
		db = newPGDB(opts.Tx, s.factory)
	} else {
		db = newPGDB(s.DB, s.factory)
	}

	// lock webhooks table to prevent concurrent updates (a row-level lock is
	// insufficient)
	err := db.lock(ctx, func(tx *pgdb) error {
		hookID, vcsProviderID, err := tx.deleteConnection(ctx, opts)
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
		client, err := s.GetVCSClient(ctx, vcsProviderID)
		if err != nil {
			return err
		}
		err = client.DeleteWebhook(ctx, cloud.DeleteWebhookOptions{
			Repo: hook.identifier,
			ID:   *hook.cloudID,
		})
		if err != nil {
			s.Error(err, "deleting webhook", "repo", hook.identifier, "cloud", hook.cloud)
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
