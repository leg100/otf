package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	db struct {
		*sql.DB
		factory
	}

	hookRow struct {
		WebhookID     pgtype.UUID `json:"webhook_id"`
		VCSID         pgtype.Text `json:"vcs_id"`
		VCSProviderID pgtype.Text `json:"vcs_provider_id"`
		Secret        pgtype.Text `json:"secret"`
		Identifier    pgtype.Text `json:"identifier"`
		Cloud         pgtype.Text `json:"cloud"`
	}
)

// GetByID implements pubsub.Getter
func (db *db) GetByID(ctx context.Context, rawID string, action pubsub.DBAction) (any, error) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, err
	}
	if action == pubsub.DeleteDBAction {
		return &hook{id: id}, nil
	}
	return db.getHookByID(ctx, id)
}

// getOrCreate gets a hook if it exists or creates it if it does not. Should be
// called within a tx to avoid concurrent access causing unpredictible results.
func (db *db) getOrCreateHook(ctx context.Context, hook *hook) (*hook, error) {
	q := db.Conn(ctx)
	result, err := q.FindWebhookByRepo(ctx, pggen.FindWebhookByRepoParams{
		Identifier:    sql.String(hook.identifier),
		VCSProviderID: sql.String(hook.vcsProviderID),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	if len(result) > 0 {
		return db.fromRow(hookRow(result[0]))
	}

	// not found; create instead

	insertResult, err := q.InsertWebhook(ctx, pggen.InsertWebhookParams{
		WebhookID:     sql.UUID(hook.id),
		Secret:        sql.String(hook.secret),
		Identifier:    sql.String(hook.identifier),
		VCSID:         sql.StringPtr(hook.cloudID),
		VCSProviderID: sql.String(hook.vcsProviderID),
	})
	if err != nil {
		return nil, fmt.Errorf("inserting webhook into db: %w", sql.Error(err))
	}
	return db.fromRow(hookRow(insertResult))
}

func (db *db) getHookByID(ctx context.Context, id uuid.UUID) (*hook, error) {
	q := db.Conn(ctx)
	result, err := q.FindWebhookByID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.fromRow(hookRow(result))
}

func (db *db) listHooks(ctx context.Context) ([]*hook, error) {
	q := db.Conn(ctx)
	result, err := q.FindWebhooks(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	hooks := make([]*hook, len(result))
	for i, row := range result {
		hook, err := db.fromRow(hookRow(row))
		if err != nil {
			return nil, sql.Error(err)
		}
		hooks[i] = hook
	}
	return hooks, nil
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
		WebhookID: sql.UUID(hookID),
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

func (db *db) deleteConnection(ctx context.Context, opts DisconnectOptions) (err error) {
	q := db.Conn(ctx)
	switch opts.ConnectionType {
	case WorkspaceConnection:
		_, err = q.DeleteWorkspaceConnectionByID(ctx, sql.String(opts.ResourceID))
	case ModuleConnection:
		_, err = q.DeleteModuleConnectionByID(ctx, sql.String(opts.ResourceID))
	default:
		return fmt.Errorf("unknown connection type: %v", opts.ConnectionType)
	}
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
