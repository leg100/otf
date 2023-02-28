package hooks

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
)

// RepoDeleteError is a non-fatal error from attempting to remove a webhook from
// a repository.
type RepoDeleteError struct {
	err error
}

func (e *RepoDeleteError) Is(err error) bool { return err == otf.ErrWarning }

type Service struct {
	db      // access to hook database
	factory // produce new hooks
}

func NewService(opts NewServiceOptions) *Service {
	factory := newFactory(opts.HostnameService, opts.CloudService)
	return &Service{
		db:      newPGDB(opts.Database, factory),
		factory: factory,
	}
}

type NewServiceOptions struct {
	CloudService cloud.Service

	otf.Database
	otf.HostnameService
}

// Hook hooks up a resource to a VCS repository, so that it may subscribe to its VCS
// events. A webhook is configured on the repo if one doesn't exist already. The
// caller provides a callback with which to establish a relationship in the DB
// between a resource and the hook.
func (s *Service) Hook(ctx context.Context, opts otf.HookOptions) error {
	unsynced, err := s.newHook(newHookOpts{
		identifier: opts.Identifier,
		cloud:      opts.Cloud,
	})
	if err != nil {
		return err
	}

	return s.tx(ctx, func(tx db) error {
		// create or get hook from db, and synchronise.
		synced, err := tx.create(ctx, unsynced, opts.Client)
		if err != nil {
			return err
		}
		// connect resource to hook
		return opts.HookCallback(ctx, tx, synced.id)
	})
}

// Unhook unhooks a resource from a VCS repository. If no other resources are
// hooked up then the webhook is deleted from the repo. The caller provides a
// callback with which to remove the relationship between the hook and the
// resource in the DB.
//
// NOTE: if the webhook cannot be deleted from the repo then this is not deemed
// fatal and the hook is still deleted from the database.
func (s *Service) Unhook(ctx context.Context, opts otf.UnhookOptions) error {
	// separately capture any error resulting from attempting to delete the
	// webhook from the repo
	var repoErr error

	txErr := s.tx(ctx, func(tx db) error {
		// disconnect connected resource
		if err := opts.UnhookCallback(ctx, tx); err != nil {
			return err
		}

		// remove hook from DB
		hook, err := tx.delete(ctx, opts.HookID)
		if err == errConnected {
			// other resources still connected
			return nil
		} else if err != nil {
			return err
		}

		// remove hook from repo
		err = opts.DeleteWebhook(ctx, cloud.DeleteWebhookOptions{
			Identifier: hook.identifier,
			ID:         *hook.cloudID,
		})
		if err != nil {
			repoErr = fmt.Errorf("%w: unable to delete webhook from repo: %w", otf.ErrWarning, err)
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}
	return repoErr
}
