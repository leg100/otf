package repohooks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/vcs"
)

type (
	db struct {
		*sql.DB
		*internal.HostnameService
	}

	hookRow struct {
		RepohookID    pgtype.UUID `json:"repohook_id"`
		VCSID         pgtype.Text `json:"vcs_id"`
		VCSProviderID pgtype.Text `json:"vcs_provider_id"`
		Secret        pgtype.Text `json:"secret"`
		RepoPath      pgtype.Text `json:"repo_path"`
		VCSKind       pgtype.Text `json:"vcs_kind"`
	}
)

// getOrCreateHook gets a hook if it exists or creates it if it does not. Should be
// called within a tx to avoid concurrent access causing unpredictible results.
func (db *db) getOrCreateHook(ctx context.Context, hook *hook) (*hook, error) {
	q := db.Querier(ctx)
	result, err := q.FindRepohookByRepoAndProvider(ctx, sqlc.FindRepohookByRepoAndProviderParams{
		RepoPath:      sql.String(hook.repoPath),
		VCSProviderID: sql.String(hook.vcsProviderID),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	if len(result) > 0 {
		return db.fromRow(hookRow(result[0]))
	}

	// not found; create instead

	insertResult, err := q.InsertRepohook(ctx, sqlc.InsertRepohookParams{
		RepohookID:    sql.UUID(hook.id),
		Secret:        sql.String(hook.secret),
		RepoPath:      sql.String(hook.repoPath),
		VCSID:         sql.StringPtr(hook.cloudID),
		VCSProviderID: sql.String(hook.vcsProviderID),
	})
	if err != nil {
		return nil, fmt.Errorf("inserting webhook into db: %w", sql.Error(err))
	}
	return db.fromRow(hookRow(insertResult))
}

func (db *db) getHookByID(ctx context.Context, id uuid.UUID) (*hook, error) {
	q := db.Querier(ctx)
	result, err := q.FindRepohookByID(ctx, sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.fromRow(hookRow(result))
}

func (db *db) listHooks(ctx context.Context) ([]*hook, error) {
	q := db.Querier(ctx)
	result, err := q.FindRepohooks(ctx)
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

func (db *db) listUnreferencedRepohooks(ctx context.Context) ([]*hook, error) {
	q := db.Querier(ctx)
	result, err := q.FindUnreferencedRepohooks(ctx)
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
	q := db.Querier(ctx)
	_, err := q.UpdateRepohookVCSID(ctx, sqlc.UpdateRepohookVCSIDParams{
		VCSID:      sql.String(cloudID),
		RepohookID: sql.UUID(id),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) deleteHook(ctx context.Context, id uuid.UUID) error {
	q := db.Querier(ctx)
	_, err := q.DeleteRepohookByID(ctx, sql.UUID(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// fromRow creates a hook from a database row
func (db *db) fromRow(row hookRow) (*hook, error) {
	opts := newRepohookOptions{
		id:              internal.UUID(row.RepohookID.Bytes),
		vcsProviderID:   row.VCSProviderID.String,
		secret:          internal.String(row.Secret.String),
		repoPath:        row.RepoPath.String,
		cloud:           vcs.Kind(row.VCSKind.String),
		HostnameService: db.HostnameService,
	}
	if row.VCSID.Valid {
		opts.cloudID = internal.String(row.VCSID.String)
	}
	return newRepohook(opts)
}
