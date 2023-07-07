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
	db struct {
		*sql.DB
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

func (db *db) unmarshal(row hookRow) (*hook, error) {
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

// getOrCreate gets a hook if it exists or creates it if it does not. Should be
// called within a tx to avoid concurrent access causing unpredictible results.
func (db *db) getOrCreateHook(ctx context.Context, hook *hook) (*hook, error) {
	q := db.Conn(ctx)
	result, err := q.FindWebhookByRepo(ctx, sql.String(hook.identifier), sql.String(hook.cloud))
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
	insertResult, err := q.InsertWebhook(ctx, params)
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(hookRow(insertResult))
}

func (db *db) getHookByID(ctx context.Context, id uuid.UUID) (*hook, error) {
	q := db.Conn(ctx)
	result, err := q.FindWebhookByID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(hookRow(result))
}

func (db *db) getHookByIDForUpdate(ctx context.Context, id uuid.UUID) (*hook, error) {
	q := db.Conn(ctx)
	result, err := q.FindWebhookByIDForUpdate(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(hookRow(result))
}

func (db *db) updateHookCloudID(ctx context.Context, id uuid.UUID, cloudID string) error {
	q := db.Conn(ctx)
	_, err := q.UpdateWebhookVCSID(ctx, sql.String(cloudID), sql.UUID(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) createConnection(ctx context.Context, hookID uuid.UUID, opts ConnectOptions) error {
	q := db.Conn(ctx)
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

	if _, err := q.InsertRepoConnection(ctx, params); err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) deleteConnection(ctx context.Context, opts DisconnectOptions) (hookID uuid.UUID, vcsProviderID string, err error) {
	q := db.Conn(ctx)
	switch opts.ConnectionType {
	case WorkspaceConnection:
		result, err := q.DeleteWorkspaceConnectionByID(ctx, sql.String(opts.ResourceID))
		if err != nil {
			return uuid.UUID{}, "", sql.Error(err)
		}
		return result.WebhookID.Bytes, result.VCSProviderID.String, nil
	case ModuleConnection:
		result, err := q.DeleteModuleConnectionByID(ctx, sql.String(opts.ResourceID))
		if err != nil {
			return uuid.UUID{}, "", sql.Error(err)
		}
		return result.WebhookID.Bytes, result.VCSProviderID.String, nil
	default:
		return uuid.UUID{}, "", fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
	}
}

func (db *db) countConnections(ctx context.Context, hookID uuid.UUID) (int, error) {
	q := db.Conn(ctx)
	result, err := q.CountRepoConnectionsByID(ctx, sql.UUID(hookID))
	if err != nil {
		return 0, err
	}
	return int(result.Int), nil
}

func (db *db) deleteHook(ctx context.Context, id uuid.UUID) (*hook, error) {
	q := db.Conn(ctx)
	result, err := q.DeleteWebhookByID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(hookRow(result))
}
