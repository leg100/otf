package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	db struct {
		*sql.DB
		internal.HostnameService
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
	result, err := q.FindWebhookByRepoAndProvider(ctx, sql.String(hook.identifier), sql.String(hook.vcsProviderID))
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

func (db *db) listUnreferencedWebhooks(ctx context.Context) ([]*hook, error) {
	q := db.Conn(ctx)
	result, err := q.FindUnreferencedWebhooks(ctx)
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

func (db *db) deleteHook(ctx context.Context, id uuid.UUID) error {
	q := db.Conn(ctx)
	_, err := q.DeleteWebhookByID(ctx, sql.UUID(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// fromRow creates a hook from a database row
func (db *db) fromRow(row hookRow) (*hook, error) {
	opts := newHookOptions{
		id:              internal.UUID(row.WebhookID.Bytes),
		vcsProviderID:   row.VCSProviderID.String,
		secret:          internal.String(row.Secret.String),
		identifier:      row.Identifier.String,
		cloud:           cloud.Kind(row.Cloud.String),
		HostnameService: db.HostnameService,
	}
	if row.VCSID.Status == pgtype.Present {
		opts.cloudID = internal.String(row.VCSID.String)
	}
	return newHook(opts)
}
