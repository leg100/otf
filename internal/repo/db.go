package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is the repo database on postgres
	pgdb struct {
		internal.DB
		factory
	}
	hookRow struct {
		WebhookID  pgtype.UUID `json:"webhook_id"`
		VCSID      pgtype.Text `json:"vcs_id"`
		Secret     pgtype.Text `json:"secret"`
		Identifier pgtype.Text `json:"identifier"`
		Cloud      pgtype.Text `json:"cloud"`
	}
)

func newPGDB(db internal.DB, f factory) *pgdb {
	return &pgdb{db, f}
}

// getOrCreate gets a hook if it exists or creates it if it does not. Should be
// called within a tx to avoid concurrent access causing unpredictible results.
func (db *pgdb) getOrCreateHook(ctx context.Context, hook *hook) (*hook, error) {
	result, err := db.FindWebhookByRepo(ctx, sql.String(hook.identifier), sql.String(hook.cloud))
	if err != nil {
		return nil, sql.Error(err)
	}
	if len(result) > 0 {
		return db.unmarshal(hookRow(result[0]))
	}

	// not found; create instead

	params := pggen.InsertWebhookParams{
		WebhookID:  sql.UUID(hook.id),
		Secret:     sql.String(hook.secret),
		Identifier: sql.String(hook.identifier),
		Cloud:      sql.String(hook.cloud),
	}
	if hook.cloudID != nil {
		params.VCSID = sql.String(*hook.cloudID)
	} else {
		params.VCSID = pgtype.Text{Status: pgtype.Null}
	}
	insertResult, err := db.InsertWebhook(ctx, params)
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(hookRow(insertResult))
}

func (db *pgdb) getHookByID(ctx context.Context, id uuid.UUID) (*hook, error) {
	result, err := db.FindWebhookByID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(hookRow(result))
}

func (db *pgdb) getHookByIDForUpdate(ctx context.Context, id uuid.UUID) (*hook, error) {
	result, err := db.FindWebhookByIDForUpdate(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(hookRow(result))
}

func (db *pgdb) updateHookCloudID(ctx context.Context, id uuid.UUID, cloudID string) error {
	_, err := db.UpdateWebhookVCSID(ctx, sql.String(cloudID), sql.UUID(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) createConnection(ctx context.Context, hookID uuid.UUID, opts ConnectOptions) error {
	params := pggen.InsertRepoConnectionParams{
		WebhookID:     sql.UUID(hookID),
		VCSProviderID: sql.String(opts.VCSProviderID),
	}

	switch opts.ConnectionType {
	case WorkspaceConnection:
		params.WorkspaceID = sql.String(opts.ResourceID)
		params.ModuleID = pgtype.Text{Status: pgtype.Null}
	case ModuleConnection:
		params.ModuleID = sql.String(opts.ResourceID)
		params.WorkspaceID = pgtype.Text{Status: pgtype.Null}
	default:
		return fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
	}

	if _, err := db.InsertRepoConnection(ctx, params); err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) deleteConnection(ctx context.Context, opts DisconnectOptions) (hookID uuid.UUID, vcsProviderID string, err error) {
	switch opts.ConnectionType {
	case WorkspaceConnection:
		result, err := db.DeleteWorkspaceConnectionByID(ctx, sql.String(opts.ResourceID))
		if err != nil {
			return uuid.UUID{}, "", sql.Error(err)
		}
		return result.WebhookID.Bytes, result.VCSProviderID.String, nil
	case ModuleConnection:
		result, err := db.DeleteModuleConnectionByID(ctx, sql.String(opts.ResourceID))
		if err != nil {
			return uuid.UUID{}, "", sql.Error(err)
		}
		return result.WebhookID.Bytes, result.VCSProviderID.String, nil
	default:
		return uuid.UUID{}, "", fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
	}
}

func (db *pgdb) countConnections(ctx context.Context, hookID uuid.UUID) (int, error) {
	return db.CountRepoConnectionsByID(ctx, sql.UUID(hookID))
}

func (db *pgdb) deleteHook(ctx context.Context, id uuid.UUID) (*hook, error) {
	result, err := db.DeleteWebhookByID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(hookRow(result))
}

// lock webhooks table within a transaction, preventing anything else from
// updating the table. Provides a callback within which caller can use the
// transaction.
func (db *pgdb) lock(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx internal.DB) error {
		if _, err := tx.Exec(ctx, "LOCK TABLE tags IN EXCLUSIVE MODE"); err != nil {
			return err
		}
		return callback(newPGDB(tx, db.factory))
	})
}

func (db *pgdb) unmarshal(row hookRow) (*hook, error) {
	opts := newHookOpts{
		id:         internal.UUID(row.WebhookID.Bytes),
		secret:     internal.String(row.Secret.String),
		identifier: row.Identifier.String,
		cloud:      row.Cloud.String,
	}
	if row.VCSID.Status == pgtype.Present {
		opts.cloudID = internal.String(row.VCSID.String)
	}
	return db.newHook(opts)
}
