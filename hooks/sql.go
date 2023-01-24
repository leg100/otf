package hooks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a hook database
type db interface {
	otf.Database

	create(context.Context, *hook, cloud.Client) (*hook, error)
	get(context.Context, uuid.UUID) (*hook, error)
	delete(context.Context, uuid.UUID) (*hook, error)
	tx(context.Context, func(db) error) error
}

// pgdb is the hook database on postgres
type pgdb struct {
	otf.Database
	factory
}

func newPGDB(db otf.Database, f factory) *pgdb {
	return &pgdb{db, f}
}

// create idempotently persists a hook to the db in a three-step synchronisation process:
// 1) create hook or get existing hook from db
// 2) invoke callback with hook, which returns the cloud provider's hook ID
// 3) update hook in DB with the ID
func (db *pgdb) create(ctx context.Context, unsynced *hook, client cloud.Client) (*hook, error) {
	var hook *hook
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
		unsynced, err = db.unmarshal(row(upsertResult))
		if err != nil {
			return err
		}
		if err := unsynced.sync(ctx, client); err != nil {
			return err
		}
		updateResult, err := tx.UpdateWebhookVCSID(ctx, sql.String(*unsynced.cloudID), upsertResult.WebhookID)
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

func (db *pgdb) get(ctx context.Context, id uuid.UUID) (*hook, error) {
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
func (db *pgdb) delete(ctx context.Context, id uuid.UUID) (*hook, error) {
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

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(db) error) error {
	return db.Transaction(ctx, func(tx otf.Database) error {
		return callback(newPGDB(tx, db.factory))
	})
}

type row struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
	Connected  int         `json:"connected"`
}

func (db *pgdb) unmarshal(r row) (*hook, error) {
	opts := newHookOpts{
		id:         otf.UUID(r.WebhookID.Bytes),
		secret:     otf.String(r.Secret.String),
		identifier: r.Identifier.String,
		cloud:      r.Cloud.String,
	}
	if r.VCSID.Status == pgtype.Present {
		opts.cloudID = otf.String(r.VCSID.String)
	}
	return db.newHook(opts)
}
