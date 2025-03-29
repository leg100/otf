package repohooks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
)

type db struct {
	*sql.DB
	*internal.HostnameService
}

// getOrCreateHook gets a hook if it exists or creates it if it does not. Should be
// called within a tx to avoid concurrent access causing unpredictible results.
func (db *db) getOrCreateHook(ctx context.Context, hook *hook) (*hook, error) {
	// Credit to https://stackoverflow.com/a/62205017
	rows := db.Query(ctx, `
WITH inserted AS (
    INSERT INTO repohooks (
        repohook_id,
        vcs_id,
        vcs_provider_id,
        secret,
        repo_path
    ) VALUES (
        @repohook_id,
        @vcs_id,
        @vcs_provider_id,
        @secret,
        @repo_path
    )
	ON CONFLICT DO NOTHING
	RETURNING *
)
SELECT i.*, v.vcs_kind
FROM inserted i
JOIN vcs_providers v USING (vcs_provider_id)
UNION
SELECT h.*, v.vcs_kind
FROM repohooks h
JOIN vcs_providers v USING (vcs_provider_id)
WHERE h.repo_path = @repo_path
AND   h.vcs_provider_id = @vcs_provider_id
`, pgx.NamedArgs{
		"repohook_id":     hook.id,
		"repo_path":       hook.repoPath,
		"vcs_id":          hook.cloudID,
		"secret":          hook.secret,
		"vcs_provider_id": hook.vcsProviderID,
	})
	return sql.CollectOneRow(rows, db.scan)
}

func (db *db) getHookByID(ctx context.Context, id uuid.UUID) (*hook, error) {
	rows := db.Query(ctx, `
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE w.repohook_id = $1
`, id)
	return sql.CollectOneRow(rows, db.scan)
}

func (db *db) listHooks(ctx context.Context) ([]*hook, error) {
	rows := db.Query(ctx, `
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
`)
	return sql.CollectRows(rows, db.scan)
}

func (db *db) listUnreferencedRepohooks(ctx context.Context) ([]*hook, error) {
	rows := db.Query(ctx, `
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE NOT EXISTS (
    SELECT FROM repo_connections rc
    WHERE rc.vcs_provider_id = w.vcs_provider_id
    AND   rc.repo_path = w.repo_path
)`)
	return sql.CollectRows(rows, db.scan)
}

func (db *db) updateHookCloudID(ctx context.Context, id uuid.UUID, cloudID string) error {
	_, err := db.Exec(ctx, `
UPDATE repohooks
SET vcs_id = $1
WHERE repohook_id = $2
RETURNING repohook_id, vcs_id, secret, repo_path, vcs_provider_id
`, cloudID, id)
	return err
}

func (db *db) deleteHook(ctx context.Context, id uuid.UUID) error {
	_, err := db.Exec(ctx,
		`
DELETE
FROM repohooks
WHERE repohook_id = $1
RETURNING repohook_id, vcs_id, secret, repo_path, vcs_provider_id
`, id)
	if err != nil {
		return err
	}
	return nil
}

type hookModel struct {
	RepohookID    uuid.UUID      `db:"repohook_id"`
	VCSID         *string        `db:"vcs_id"`
	VCSProviderID resource.TfeID `db:"vcs_provider_id"`
	Secret        string         `db:"secret"`
	RepoPath      string         `db:"repo_path"`
	VCSKind       vcs.Kind       `db:"vcs_kind"`
}

// fromRow creates a hook from a database row
func (db *db) scan(row pgx.CollectableRow) (*hook, error) {
	model, err := pgx.RowToStructByName[hookModel](row)
	if err != nil {
		return nil, err
	}
	opts := newRepohookOptions{
		id:              &model.RepohookID,
		vcsProviderID:   model.VCSProviderID,
		secret:          &model.Secret,
		repoPath:        model.RepoPath,
		cloud:           model.VCSKind,
		HostnameService: db.HostnameService,
		cloudID:         model.VCSID,
	}
	return newRepohook(opts)
}
