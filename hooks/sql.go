package hooks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type pgdb struct {
	otf.Database
	cloud.Service
	otf.HostnameService
}

// create idempotently persists a hook to the db in a three-step synchronisation process:
// 1) create hook or get existing hook from db
// 2) invoke callback with hook, which returns the cloud provider's hook ID
// 3) update hook in DB with the ID
func (db *pgdb) create(ctx context.Context, unsynced *unsynced, fn syncFunc) (*synced, error) {
	var hook *synced
	err := db.Transaction(ctx, func(tx otf.Database) error {
		upsertResult, err := tx.UpsertWebhook(ctx, pggen.UpsertWebhookParams{
			WebhookID:  sql.UUID(unsynced.id),
			Secret:     sql.String(unsynced.secret),
			Identifier: sql.String(unsynced.identifier),
			Cloud:      sql.String(unsynced.cloud),
		})
		if err != nil {
			return err
		}
		if upsertResult.VCSID.Status == pgtype.Present {
			hook, err = db.unmarshal(row(upsertResult))
			if err != nil {
				return err
			}
		}
		cloudID, err := fn(hook, tx)
		if err != nil {
			return err
		}
		updateResult, err := tx.UpdateWebhookVCSID(ctx, sql.String(cloudID), upsertResult.WebhookID)
		if err != nil {
			return err
		}
		hook, err = db.unmarshal(row(updateResult))
		if err != nil {
			return err
		}
		return err
	})
	return hook, sql.Error(err)
}

func (db *pgdb) get(ctx context.Context, id uuid.UUID) (*synced, error) {
	result, err := db.FindWebhookByID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	hook, err := db.unmarshal(row(result))
	if err != nil {
		return nil, err
	}
	return hook, nil
}

// delete deletes the webhook from the database but not before it
// decrements the number of 'connections' to the webhook and only if the number
// is zero does it delete the webhook; otherwise it returns
// errConnected.
func (db *pgdb) delete(ctx context.Context, id uuid.UUID) (*synced, error) {
	var r pggen.DisconnectWebhookRow
	err := db.Transaction(ctx, func(db otf.Database) (err error) {
		r, err = db.DisconnectWebhook(ctx, sql.UUID(id))
		if err != nil {
			return err
		}
		if r.Connected > 0 {
			return nil
		}
		_, err = db.DeleteWebhookByID(ctx, sql.UUID(id))
		return err
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	if r.Connected > 0 {
		return nil, errConnected
	}
	return db.unmarshal(row(r))
}

type row struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
	Connected  int         `json:"connected"`
}

func (db *pgdb) unmarshal(r row) (*synced, error) {
	cloudConfig, err := db.GetCloudConfig(r.Cloud.String)
	if err != nil {
		return nil, fmt.Errorf("unknown cloud: %s", cloudConfig)
	}
	hook, err := newHook(newHookOpts{
		id:         otf.UUID(r.WebhookID.Bytes),
		secret:     otf.String(r.Secret.String),
		identifier: r.Identifier.String,
		cloud:      r.Cloud.String,
		hostname:   db.Hostname(),
	})
	if err != nil {
		return nil, err
	}
	return &synced{
		unsynced:     hook,
		EventHandler: cloudConfig,
		cloudID:      r.VCSID.String,
	}, nil
}
