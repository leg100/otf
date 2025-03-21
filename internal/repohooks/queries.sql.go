// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queries.sql

package repohooks

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

const deleteRepohookByID = `-- name: DeleteRepohookByID :one
DELETE
FROM repohooks
WHERE repohook_id = $1
RETURNING repohook_id, vcs_id, secret, repo_path, vcs_provider_id
`

func (q *Queries) DeleteRepohookByID(ctx context.Context, db DBTX, repohookID pgtype.UUID) (Repohook, error) {
	row := db.QueryRow(ctx, deleteRepohookByID, repohookID)
	var i Repohook
	err := row.Scan(
		&i.RepohookID,
		&i.VCSID,
		&i.Secret,
		&i.RepoPath,
		&i.VCSProviderID,
	)
	return i, err
}

const findRepohookByID = `-- name: FindRepohookByID :one
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
`

type FindRepohookByIDRow struct {
	RepohookID    pgtype.UUID
	VCSID         pgtype.Text
	VCSProviderID resource.TfeID
	Secret        pgtype.Text
	RepoPath      pgtype.Text
	VCSKind       pgtype.Text
}

func (q *Queries) FindRepohookByID(ctx context.Context, db DBTX, repohookID pgtype.UUID) (FindRepohookByIDRow, error) {
	row := db.QueryRow(ctx, findRepohookByID, repohookID)
	var i FindRepohookByIDRow
	err := row.Scan(
		&i.RepohookID,
		&i.VCSID,
		&i.VCSProviderID,
		&i.Secret,
		&i.RepoPath,
		&i.VCSKind,
	)
	return i, err
}

const findRepohookByRepoAndProvider = `-- name: FindRepohookByRepoAndProvider :many
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
WHERE repo_path = $1
AND   w.vcs_provider_id = $2
`

type FindRepohookByRepoAndProviderParams struct {
	RepoPath      pgtype.Text
	VCSProviderID resource.TfeID
}

type FindRepohookByRepoAndProviderRow struct {
	RepohookID    pgtype.UUID
	VCSID         pgtype.Text
	VCSProviderID resource.TfeID
	Secret        pgtype.Text
	RepoPath      pgtype.Text
	VCSKind       pgtype.Text
}

func (q *Queries) FindRepohookByRepoAndProvider(ctx context.Context, db DBTX, arg FindRepohookByRepoAndProviderParams) ([]FindRepohookByRepoAndProviderRow, error) {
	rows, err := db.Query(ctx, findRepohookByRepoAndProvider, arg.RepoPath, arg.VCSProviderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindRepohookByRepoAndProviderRow
	for rows.Next() {
		var i FindRepohookByRepoAndProviderRow
		if err := rows.Scan(
			&i.RepohookID,
			&i.VCSID,
			&i.VCSProviderID,
			&i.Secret,
			&i.RepoPath,
			&i.VCSKind,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findRepohooks = `-- name: FindRepohooks :many
SELECT
    w.repohook_id,
    w.vcs_id,
    w.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM repohooks w
JOIN vcs_providers v USING (vcs_provider_id)
`

type FindRepohooksRow struct {
	RepohookID    pgtype.UUID
	VCSID         pgtype.Text
	VCSProviderID resource.TfeID
	Secret        pgtype.Text
	RepoPath      pgtype.Text
	VCSKind       pgtype.Text
}

func (q *Queries) FindRepohooks(ctx context.Context, db DBTX) ([]FindRepohooksRow, error) {
	rows, err := db.Query(ctx, findRepohooks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindRepohooksRow
	for rows.Next() {
		var i FindRepohooksRow
		if err := rows.Scan(
			&i.RepohookID,
			&i.VCSID,
			&i.VCSProviderID,
			&i.Secret,
			&i.RepoPath,
			&i.VCSKind,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findUnreferencedRepohooks = `-- name: FindUnreferencedRepohooks :many
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
)
`

type FindUnreferencedRepohooksRow struct {
	RepohookID    pgtype.UUID
	VCSID         pgtype.Text
	VCSProviderID resource.TfeID
	Secret        pgtype.Text
	RepoPath      pgtype.Text
	VCSKind       pgtype.Text
}

func (q *Queries) FindUnreferencedRepohooks(ctx context.Context, db DBTX) ([]FindUnreferencedRepohooksRow, error) {
	rows, err := db.Query(ctx, findUnreferencedRepohooks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindUnreferencedRepohooksRow
	for rows.Next() {
		var i FindUnreferencedRepohooksRow
		if err := rows.Scan(
			&i.RepohookID,
			&i.VCSID,
			&i.VCSProviderID,
			&i.Secret,
			&i.RepoPath,
			&i.VCSKind,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertRepohook = `-- name: InsertRepohook :one
WITH inserted AS (
    INSERT INTO repohooks (
        repohook_id,
        vcs_id,
        vcs_provider_id,
        secret,
        repo_path
    ) VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING repohook_id, vcs_id, secret, repo_path, vcs_provider_id
)
SELECT
    w.repohook_id,
    w.vcs_id,
    v.vcs_provider_id,
    w.secret,
    w.repo_path,
    v.vcs_kind
FROM inserted w
JOIN vcs_providers v USING (vcs_provider_id)
`

type InsertRepohookParams struct {
	RepohookID    pgtype.UUID
	VCSID         pgtype.Text
	VCSProviderID resource.TfeID
	Secret        pgtype.Text
	RepoPath      pgtype.Text
}

type InsertRepohookRow struct {
	RepohookID    pgtype.UUID
	VCSID         pgtype.Text
	VCSProviderID resource.TfeID
	Secret        pgtype.Text
	RepoPath      pgtype.Text
	VCSKind       pgtype.Text
}

func (q *Queries) InsertRepohook(ctx context.Context, db DBTX, arg InsertRepohookParams) (InsertRepohookRow, error) {
	row := db.QueryRow(ctx, insertRepohook,
		arg.RepohookID,
		arg.VCSID,
		arg.VCSProviderID,
		arg.Secret,
		arg.RepoPath,
	)
	var i InsertRepohookRow
	err := row.Scan(
		&i.RepohookID,
		&i.VCSID,
		&i.VCSProviderID,
		&i.Secret,
		&i.RepoPath,
		&i.VCSKind,
	)
	return i, err
}

const updateRepohookVCSID = `-- name: UpdateRepohookVCSID :one
UPDATE repohooks
SET vcs_id = $1
WHERE repohook_id = $2
RETURNING repohook_id, vcs_id, secret, repo_path, vcs_provider_id
`

type UpdateRepohookVCSIDParams struct {
	VCSID      pgtype.Text
	RepohookID pgtype.UUID
}

func (q *Queries) UpdateRepohookVCSID(ctx context.Context, db DBTX, arg UpdateRepohookVCSIDParams) (Repohook, error) {
	row := db.QueryRow(ctx, updateRepohookVCSID, arg.VCSID, arg.RepohookID)
	var i Repohook
	err := row.Scan(
		&i.RepohookID,
		&i.VCSID,
		&i.Secret,
		&i.RepoPath,
		&i.VCSProviderID,
	)
	return i, err
}
