package repohooks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
)

var q = &Queries{}

type (
	db struct {
		*sql.DB
		*internal.HostnameService
	}

	hookRow struct {
		RepohookID    pgtype.UUID `json:"repohook_id"`
		VCSID         pgtype.Text `json:"vcs_id"`
		VCSProviderID resource.ID `json:"vcs_provider_id"`
		Secret        pgtype.Text `json:"secret"`
		RepoPath      pgtype.Text `json:"repo_path"`
		VCSKind       pgtype.Text `json:"vcs_kind"`
	}
)

// getOrCreateHook gets a hook if it exists or creates it if it does not. Should be
// called within a tx to avoid concurrent access causing unpredictible results.
func (db *db) getOrCreateHook(ctx context.Context, hook *hook) (*hook, error) {
	result, err := q.FindRepohookByRepoAndProvider(ctx, db.Conn(ctx), FindRepohookByRepoAndProviderParams{
		RepoPath:      sql.String(hook.repoPath),
		VCSProviderID: hook.vcsProviderID,
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	if len(result) > 0 {
		return db.fromRow(hookRow(result[0]))
	}

	// not found; create instead

	insertResult, err := q.InsertRepohook(ctx, db.Conn(ctx), InsertRepohookParams{
		RepohookID:    sql.UUID(hook.id),
		Secret:        sql.String(hook.secret),
		RepoPath:      sql.String(hook.repoPath),
		VCSID:         sql.StringPtr(hook.cloudID),
		VCSProviderID: hook.vcsProviderID,
	})
	if err != nil {
		return nil, fmt.Errorf("inserting webhook into db: %w", sql.Error(err))
	}
	return db.fromRow(hookRow(insertResult))
}

func (db *db) getHookByID(ctx context.Context, id uuid.UUID) (*hook, error) {
	result, err := q.FindRepohookByID(ctx, db.Conn(ctx), sql.UUID(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.fromRow(hookRow(result))
}

func (db *db) listHooks(ctx context.Context) ([]*hook, error) {
	result, err := q.FindRepohooks(ctx, db.Conn(ctx))
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
	result, err := q.FindUnreferencedRepohooks(ctx, db.Conn(ctx))
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
	_, err := q.UpdateRepohookVCSID(ctx, db.Conn(ctx), UpdateRepohookVCSIDParams{
		VCSID:      sql.String(cloudID),
		RepohookID: sql.UUID(id),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *db) deleteHook(ctx context.Context, id uuid.UUID) error {
	_, err := q.DeleteRepohookByID(ctx, db.Conn(ctx), sql.UUID(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// fromRow creates a hook from a database row
func (db *db) fromRow(row hookRow) (*hook, error) {
	opts := newRepohookOptions{
		id:              internal.UUID(row.RepohookID.Bytes),
		vcsProviderID:   row.VCSProviderID,
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
