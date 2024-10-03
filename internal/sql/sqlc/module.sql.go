// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: module.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteModuleByID = `-- name: DeleteModuleByID :one
DELETE
FROM modules
WHERE module_id = $1
RETURNING module_id
`

func (q *Queries) DeleteModuleByID(ctx context.Context, moduleID pgtype.Text) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, deleteModuleByID, moduleID)
	var module_id pgtype.Text
	err := row.Scan(&module_id)
	return module_id, err
}

const deleteModuleVersionByID = `-- name: DeleteModuleVersionByID :one
DELETE
FROM module_versions
WHERE module_version_id = $1
RETURNING module_version_id
`

func (q *Queries) DeleteModuleVersionByID(ctx context.Context, moduleVersionID pgtype.Text) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, deleteModuleVersionByID, moduleVersionID)
	var module_version_id pgtype.Text
	err := row.Scan(&module_version_id)
	return module_version_id, err
}

const findModuleByConnection = `-- name: FindModuleByConnection :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    array_agg(v.*)::module_versions[] AS module_versions
FROM modules m
JOIN repo_connections r USING (module_id)
LEFT JOIN module_versions v USING (module_id)
WHERE r.vcs_provider_id = $1
AND   r.repo_path = $2
GROUP BY m.module_id
`

type FindModuleByConnectionParams struct {
	VCSProviderID pgtype.Text
	RepoPath      pgtype.Text
}

type FindModuleByConnectionRow struct {
	ModuleID         pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	Provider         pgtype.Text
	Status           pgtype.Text
	OrganizationName pgtype.Text
	VCSProviderID    pgtype.Text
	RepoPath         pgtype.Text
	ModuleVersions   []ModuleVersion
}

func (q *Queries) FindModuleByConnection(ctx context.Context, arg FindModuleByConnectionParams) (FindModuleByConnectionRow, error) {
	row := q.db.QueryRow(ctx, findModuleByConnection, arg.VCSProviderID, arg.RepoPath)
	var i FindModuleByConnectionRow
	err := row.Scan(
		&i.ModuleID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Provider,
		&i.Status,
		&i.OrganizationName,
		&i.VCSProviderID,
		&i.RepoPath,
		&i.ModuleVersions,
	)
	return i, err
}

const findModuleByID = `-- name: FindModuleByID :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    array_agg(v.*)::module_versions[] AS module_versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
LEFT JOIN module_versions v USING (module_id)
WHERE m.module_id = $1
GROUP BY m.module_id
`

type FindModuleByIDRow struct {
	ModuleID         pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	Provider         pgtype.Text
	Status           pgtype.Text
	OrganizationName pgtype.Text
	VCSProviderID    pgtype.Text
	RepoPath         pgtype.Text
	ModuleVersions   []ModuleVersion
}

func (q *Queries) FindModuleByID(ctx context.Context, id pgtype.Text) (FindModuleByIDRow, error) {
	row := q.db.QueryRow(ctx, findModuleByID, id)
	var i FindModuleByIDRow
	err := row.Scan(
		&i.ModuleID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Provider,
		&i.Status,
		&i.OrganizationName,
		&i.VCSProviderID,
		&i.RepoPath,
		&i.ModuleVersions,
	)
	return i, err
}

const findModuleByModuleVersionID = `-- name: FindModuleByModuleVersionID :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    array_agg(v.*)::module_versions[] AS module_versions
FROM modules m
JOIN module_versions mv USING (module_id)
LEFT JOIN repo_connections r USING (module_id)
WHERE mv.module_version_id = $1
GROUP BY m.module_id
`

type FindModuleByModuleVersionIDRow struct {
	ModuleID         pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	Provider         pgtype.Text
	Status           pgtype.Text
	OrganizationName pgtype.Text
	VCSProviderID    pgtype.Text
	RepoPath         pgtype.Text
	ModuleVersions   []ModuleVersion
}

func (q *Queries) FindModuleByModuleVersionID(ctx context.Context, moduleVersionID pgtype.Text) (FindModuleByModuleVersionIDRow, error) {
	row := q.db.QueryRow(ctx, findModuleByModuleVersionID, moduleVersionID)
	var i FindModuleByModuleVersionIDRow
	err := row.Scan(
		&i.ModuleID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Provider,
		&i.Status,
		&i.OrganizationName,
		&i.VCSProviderID,
		&i.RepoPath,
		&i.ModuleVersions,
	)
	return i, err
}

const findModuleByName = `-- name: FindModuleByName :one
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    array_agg(v.*)::module_versions[] AS module_versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
WHERE m.organization_name = $1
AND   m.name = $2
AND   m.provider = $3
GROUP BY m.module_id
`

type FindModuleByNameParams struct {
	OrganizationName pgtype.Text
	Name             pgtype.Text
	Provider         pgtype.Text
}

type FindModuleByNameRow struct {
	ModuleID         pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	Provider         pgtype.Text
	Status           pgtype.Text
	OrganizationName pgtype.Text
	VCSProviderID    pgtype.Text
	RepoPath         pgtype.Text
	ModuleVersions   []ModuleVersion
}

func (q *Queries) FindModuleByName(ctx context.Context, arg FindModuleByNameParams) (FindModuleByNameRow, error) {
	row := q.db.QueryRow(ctx, findModuleByName, arg.OrganizationName, arg.Name, arg.Provider)
	var i FindModuleByNameRow
	err := row.Scan(
		&i.ModuleID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Provider,
		&i.Status,
		&i.OrganizationName,
		&i.VCSProviderID,
		&i.RepoPath,
		&i.ModuleVersions,
	)
	return i, err
}

const findModuleTarball = `-- name: FindModuleTarball :one
SELECT tarball
FROM module_tarballs
WHERE module_version_id = $1
`

func (q *Queries) FindModuleTarball(ctx context.Context, moduleVersionID pgtype.Text) ([]byte, error) {
	row := q.db.QueryRow(ctx, findModuleTarball, moduleVersionID)
	var tarball []byte
	err := row.Scan(&tarball)
	return tarball, err
}

const insertModule = `-- name: InsertModule :exec
INSERT INTO modules (
    module_id,
    created_at,
    updated_at,
    name,
    provider,
    status,
    organization_name
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
`

type InsertModuleParams struct {
	ID               pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	Provider         pgtype.Text
	Status           pgtype.Text
	OrganizationName pgtype.Text
}

func (q *Queries) InsertModule(ctx context.Context, arg InsertModuleParams) error {
	_, err := q.db.Exec(ctx, insertModule,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Name,
		arg.Provider,
		arg.Status,
		arg.OrganizationName,
	)
	return err
}

const insertModuleTarball = `-- name: InsertModuleTarball :one
INSERT INTO module_tarballs (
    tarball,
    module_version_id
) VALUES (
    $1,
    $2
)
RETURNING module_version_id
`

type InsertModuleTarballParams struct {
	Tarball         []byte
	ModuleVersionID pgtype.Text
}

func (q *Queries) InsertModuleTarball(ctx context.Context, arg InsertModuleTarballParams) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, insertModuleTarball, arg.Tarball, arg.ModuleVersionID)
	var module_version_id pgtype.Text
	err := row.Scan(&module_version_id)
	return module_version_id, err
}

const insertModuleVersion = `-- name: InsertModuleVersion :one
INSERT INTO module_versions (
    module_version_id,
    version,
    created_at,
    updated_at,
    module_id,
    status
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING module_version_id, version, created_at, updated_at, status, status_error, module_id
`

type InsertModuleVersionParams struct {
	ModuleVersionID pgtype.Text
	Version         pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	ModuleID        pgtype.Text
	Status          pgtype.Text
}

func (q *Queries) InsertModuleVersion(ctx context.Context, arg InsertModuleVersionParams) (ModuleVersion, error) {
	row := q.db.QueryRow(ctx, insertModuleVersion,
		arg.ModuleVersionID,
		arg.Version,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.ModuleID,
		arg.Status,
	)
	var i ModuleVersion
	err := row.Scan(
		&i.ModuleVersionID,
		&i.Version,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Status,
		&i.StatusError,
		&i.ModuleID,
	)
	return i, err
}

const listModulesByOrganization = `-- name: ListModulesByOrganization :many
SELECT
    m.module_id,
    m.created_at,
    m.updated_at,
    m.name,
    m.provider,
    m.status,
    m.organization_name,
    r.vcs_provider_id,
    r.repo_path,
    versions.module_versions
FROM modules m
LEFT JOIN repo_connections r USING (module_id)
LEFT JOIN (
    SELECT
        v.module_id,
        array_agg(v.*)::module_versions[] AS module_versions
    FROM module_versions v
    GROUP BY v.module_id
) AS versions USING (module_id)
WHERE m.organization_name = $1
`

type ListModulesByOrganizationRow struct {
	ModuleID         pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	Name             pgtype.Text
	Provider         pgtype.Text
	Status           pgtype.Text
	OrganizationName pgtype.Text
	VCSProviderID    pgtype.Text
	RepoPath         pgtype.Text
	ModuleVersions   []ModuleVersion
}

func (q *Queries) ListModulesByOrganization(ctx context.Context, organizationName pgtype.Text) ([]ListModulesByOrganizationRow, error) {
	rows, err := q.db.Query(ctx, listModulesByOrganization, organizationName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListModulesByOrganizationRow
	for rows.Next() {
		var i ListModulesByOrganizationRow
		if err := rows.Scan(
			&i.ModuleID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Provider,
			&i.Status,
			&i.OrganizationName,
			&i.VCSProviderID,
			&i.RepoPath,
			&i.ModuleVersions,
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

const updateModuleStatusByID = `-- name: UpdateModuleStatusByID :one
UPDATE modules
SET status = $1
WHERE module_id = $2
RETURNING module_id
`

type UpdateModuleStatusByIDParams struct {
	Status   pgtype.Text
	ModuleID pgtype.Text
}

func (q *Queries) UpdateModuleStatusByID(ctx context.Context, arg UpdateModuleStatusByIDParams) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, updateModuleStatusByID, arg.Status, arg.ModuleID)
	var module_id pgtype.Text
	err := row.Scan(&module_id)
	return module_id, err
}

const updateModuleVersionStatusByID = `-- name: UpdateModuleVersionStatusByID :one
UPDATE module_versions
SET
    status = $1,
    status_error = $2
WHERE module_version_id = $3
RETURNING module_version_id, version, created_at, updated_at, status, status_error, module_id
`

type UpdateModuleVersionStatusByIDParams struct {
	Status          pgtype.Text
	StatusError     pgtype.Text
	ModuleVersionID pgtype.Text
}

func (q *Queries) UpdateModuleVersionStatusByID(ctx context.Context, arg UpdateModuleVersionStatusByIDParams) (ModuleVersion, error) {
	row := q.db.QueryRow(ctx, updateModuleVersionStatusByID, arg.Status, arg.StatusError, arg.ModuleVersionID)
	var i ModuleVersion
	err := row.Scan(
		&i.ModuleVersionID,
		&i.Version,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Status,
		&i.StatusError,
		&i.ModuleID,
	)
	return i, err
}
