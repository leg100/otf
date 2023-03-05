package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is the repo database on postgres
type pgdb struct {
	*sql.DB
	factory
}

func newPGDB(db *sql.DB, f factory) *pgdb {
	return &pgdb{db, f}
}

// create idempotently persists a hook to the db in a three-step synchronisation process:
// 1) create hook or get existing hook from db
// 2) invoke callback with hook, which returns the cloud provider's hook ID
// 3) update hook in DB with the ID

func (db *pgdb) updateCloudID(ctx context.Context, id uuid.UUID, cloudID string) error {
	_, err := db.UpdateWebhookVCSID(ctx, sql.String(cloudID), sql.UUID(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// getOrCreate gets a hook if it exists or creates it if it does not. Should be
// called within a tx to avoid concurrent access causing unpredictible results.
func (db *pgdb) getOrCreate(ctx context.Context, hook *hook) (*hook, error) {
	result, err := db.FindWebhooksByRepo(ctx, sql.String(hook.identifier), sql.String(hook.cloud))
	if err != nil {
		return nil, sql.Error(err)
	}
	if len(result) > 0 {
		return db.unmarshal(pgRow(result[0]))
	}

	// not found; create instead

	insertResult, err := db.InsertWebhook(ctx, pggen.InsertWebhookParams{
		WebhookID:  sql.UUID(hook.id),
		Secret:     sql.String(hook.secret),
		Identifier: sql.String(hook.identifier),
		Cloud:      sql.String(hook.cloud),
	})
	return db.unmarshal(pgRow(insertResult))
}

func (db *pgdb) getHookByID(ctx context.Context, id uuid.UUID) (*hook, error) {
	result, err := db.FindWebhookByID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(pgRow(result))
}

func (db *pgdb) createConnection(ctx context.Context, hookID uuid.UUID, opts otf.ConnectionOptions) error {
	params := pggen.InsertRepoConnectionParams{
		WebhookID:     sql.UUID(hookID),
		VCSProviderID: sql.String(opts.VCSProviderID),
	}

	switch opts.ConnectionType {
	case otf.WorkspaceConnection:
		params.WorkspaceID = sql.String(opts.ResourceID)
		params.ModuleID = pgtype.Text{Status: pgtype.Null}
	case otf.ModuleConnection:
		params.ModuleID = sql.String(opts.ResourceID)
		params.WorkspaceID = pgtype.Text{Status: pgtype.Null}
	default:
		return fmt.Errorf("unknown connection type: %s", opts.ConnectionType)
	}

	if _, err := db.InsertRepoConnection(ctx, params); err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) deleteConnection(ctx context.Context, opts otf.ConnectionOptions) (uuid.UUID, error) {
	var hookID pgtype.UUID
	var err error

	switch opts.ConnectionType {
	case otf.WorkspaceConnection:
		hookID, err = db.DeleteWorkspaceConnectionByID(ctx, sql.String(opts.ResourceID))
	case otf.ModuleConnection:
		hookID, err = db.DeleteModuleConnectionByID(ctx, sql.String(opts.ResourceID))
	default:
		return uuid.UUID{}, fmt.Errorf("unknown connection type: %s", opts.ConnectionType)
	}
	if err != nil {
		return uuid.UUID{}, sql.Error(err)
	}
	return uuid.FromBytes(hookID.Bytes[:])
}

func (db *pgdb) countConnections(ctx context.Context, hookID uuid.UUID) (*int, error) {
	return db.CountRepoConnectionsByID(ctx, sql.UUID(hookID))
}

func (db *pgdb) deleteHook(ctx context.Context, id uuid.UUID) (*hook, error) {
	result, err := db.DeleteWebhookByID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(pgRow(result))
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) lock(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx *sql.DB) error {
		_, err := tx.Exec(ctx, "LOCK webhooks")
		if err != nil {
			return err
		}
		return callback(newPGDB(tx, db.factory))
	})
}

type pgRow struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
}

func (db *pgdb) unmarshal(row pgRow) (*hook, error) {
	opts := newHookOpts{
		id:         otf.UUID(row.WebhookID.Bytes),
		secret:     otf.String(row.Secret.String),
		identifier: row.Identifier.String,
		cloud:      row.Cloud.String,
	}
	if row.VCSID.Status == pgtype.Present {
		opts.cloudID = otf.String(row.VCSID.String)
	}
	return db.newHook(opts)
}
