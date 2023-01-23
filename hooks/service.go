package hooks

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/sql"
	"github.com/pkg/errors"
)

type Connection interface {
	Connect(tx otf.Database, hookID string) error
	Disconnect(tx otf.Database) error
}

type Service struct {
	sql.Tx
	db
	otf.HostnameService
}

type RegisterConnectionOptions struct {
	Identifier string
	Cloud      string

	Connection
	cloud.Client
}

type UnregisterConnectionOptions struct {
	HookID uuid.UUID

	Connection
	cloud.Client
}

func NewService(opts NewServiceOptions) *Service {
	return &Service{
		HostnameService: opts.HostnameService,
		db: &pgdb{
			Database:        opts.Database,
			HostnameService: opts.HostnameService,
			Service:         opts.CloudService,
		},
		Tx: opts.Database,
	}
}

type NewServiceOptions struct {
	otf.Database
	otf.HostnameService
	CloudService cloud.Service
}

// RegisterConnection establishes a connection between a VCS repository and an OTF resource,
// permitting the resource to subscribe to VCS events. The caller provides a
// connection which is called with a database transaction and ID of the hook,
// giving the caller the opportunity to register the relationship in the
// database. If there is a failure then the caller should return an error, and
// the hook will not be established.
//
// NOTE: only one hook per respository is maintained.
func (s *Service) RegisterConnection(ctx context.Context, opts RegisterConnectionOptions) error {
	unsynced, err := newHook(newHookOpts{
		identifier: opts.Identifier,
		cloud:      opts.Cloud,
		hostname:   s.Hostname(),
	})
	if err != nil {
		return err
	}

	// synchronisation function synchronises config between DB and cloud
	syncFn := func(hook *synced, tx otf.Database) (string, error) {
		if hook.cloudID == "" {
			// no hook found in DB; create on cloud
			cloudID, err := opts.CreateWebhook(ctx, cloud.CreateWebhookOptions{
				Identifier: unsynced.identifier,
				Secret:     unsynced.secret,
				Events:     defaultEvents,
				Endpoint:   unsynced.endpoint,
			})
			if err != nil {
				return "", err
			}
			return cloudID, opts.Connect(tx, unsynced.identifier)
		}

		// existing hook in DB; check it exists in cloud and create/update
		// accordingly
		cloudHook, err := opts.GetWebhook(ctx, cloud.GetWebhookOptions{
			Identifier: hook.identifier,
			ID:         hook.cloudID,
		})
		if errors.Is(err, otf.ErrResourceNotFound) {
			// hook not found in cloud; create it
			return opts.CreateWebhook(ctx, cloud.CreateWebhookOptions{
				Identifier: hook.identifier,
				Secret:     hook.secret,
				Events:     defaultEvents,
				Endpoint:   hook.endpoint,
			})
		} else if err != nil {
			return "", errors.Wrap(err, "retrieving config from cloud")
		}

		// hook found in both DB and on cloud; check for differences and update
		// accordingly
		if reflect.DeepEqual(defaultEvents, cloudHook.Events) &&
			hook.endpoint == cloudHook.Endpoint {
			// no differences
			return hook.cloudID, nil
		}

		// differences found; update hook on cloud
		err = opts.UpdateWebhook(ctx, cloud.UpdateWebhookOptions{
			ID: cloudHook.ID,
			CreateWebhookOptions: cloud.CreateWebhookOptions{
				Identifier: hook.identifier,
				Secret:     hook.secret,
				Events:     defaultEvents,
				Endpoint:   hook.endpoint,
			},
		})
		if err != nil {
			return "", err
		}

		// connect resource to hook
		return cloudHook.ID, opts.Connect(tx, hook.identifier)
	}

	// create or get hook from db, and synchronise.
	_, err = s.create(ctx, unsynced, syncFn)
	return err
}

func (s *Service) UnregisterConnection(ctx context.Context, opts UnregisterConnectionOptions) error {
	return s.Transaction(ctx, func(tx otf.Database) error {
		// disconnect connected resource
		if err := opts.Disconnect(tx); err != nil {
			return err
		}

		// remove hook from DB
		hook, err := s.delete(ctx, opts.HookID)
		if err == errConnected {
			// other resources still connected
			return nil
		} else if err != nil {
			return err
		}

		// remove hook from cloud
		return opts.DeleteWebhook(ctx, cloud.DeleteWebhookOptions{
			Identifier: hook.identifier,
			ID:         hook.cloudID,
		})
	})
}

type synchroniseOptions struct {
	cloud.Client
	proposed *unsynced
	existing *synced
}

// synchronise hook with the cloud
func synchronise(ctx context.Context, opts cloud.Client, hook *synced) (string, error) {
	if hook.cloudID == "" {
		cloudID, err := opts.CreateWebhook(ctx, cloud.CreateWebhookOptions{
			Identifier: hook.identifier,
			Secret:     hook.secret,
			Events:     defaultEvents,
			Endpoint:   hook.endpoint,
		})
		if err != nil {
			return "", err
		}
		return cloudID, nil
	}

	// existing hook in DB; check it exists in cloud and create/update
	// accordingly
	cloudHook, err := opts.GetWebhook(ctx, cloud.GetWebhookOptions{
		Identifier: hook.identifier,
		ID:         hook.cloudID,
	})
	if errors.Is(err, otf.ErrResourceNotFound) {
		// hook not found in cloud; create it
		return opts.CreateWebhook(ctx, cloud.CreateWebhookOptions{
			Identifier: hook.identifier,
			Secret:     hook.secret,
			Events:     defaultEvents,
			Endpoint:   hook.endpoint,
		})
	} else if err != nil {
		return "", errors.Wrap(err, "retrieving config from cloud")
	}

	// hook found in both DB and on cloud; check for differences and update
	// accordingly
	if reflect.DeepEqual(defaultEvents, cloudHook.Events) &&
		existing.endpoint == cloudHook.Endpoint {
		// no differences
		return existing.cloudID, nil
	}

	// differences found; update hook on cloud
	err = opts.UpdateWebhook(ctx, cloud.UpdateWebhookOptions{
		ID: cloudHook.ID,
		CreateWebhookOptions: cloud.CreateWebhookOptions{
			Identifier: existing.identifier,
			Secret:     existing.secret,
			Events:     defaultEvents,
			Endpoint:   existing.endpoint,
		},
	})
	if err != nil {
		return "", err
	}

	return cloudHook.ID, nil
}
