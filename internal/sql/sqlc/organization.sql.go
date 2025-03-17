// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: organization.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

const countOrganizations = `-- name: CountOrganizations :one
SELECT count(*)
FROM organizations
WHERE name LIKE ANY($1::text[])
`

func (q *Queries) CountOrganizations(ctx context.Context, names []pgtype.Text) (int64, error) {
	row := q.db.QueryRow(ctx, countOrganizations, names)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const deleteOrganizationByName = `-- name: DeleteOrganizationByName :one
DELETE
FROM organizations
WHERE name = $1
RETURNING organization_id
`

func (q *Queries) DeleteOrganizationByName(ctx context.Context, name organization.Name) (resource.ID, error) {
	row := q.db.QueryRow(ctx, deleteOrganizationByName, name)
	var organization_id resource.ID
	err := row.Scan(&organization_id)
	return organization_id, err
}

const findOrganizationByID = `-- name: FindOrganizationByID :one
SELECT organization_id, created_at, updated_at, name, session_remember, session_timeout, email, collaborator_auth_policy, allow_force_delete_workspaces, cost_estimation_enabled FROM organizations WHERE organization_id = $1
`

func (q *Queries) FindOrganizationByID(ctx context.Context, organizationID resource.ID) (Organization, error) {
	row := q.db.QueryRow(ctx, findOrganizationByID, organizationID)
	var i Organization
	err := row.Scan(
		&i.OrganizationID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.SessionRemember,
		&i.SessionTimeout,
		&i.Email,
		&i.CollaboratorAuthPolicy,
		&i.AllowForceDeleteWorkspaces,
		&i.CostEstimationEnabled,
	)
	return i, err
}

const findOrganizationByName = `-- name: FindOrganizationByName :one
SELECT organization_id, created_at, updated_at, name, session_remember, session_timeout, email, collaborator_auth_policy, allow_force_delete_workspaces, cost_estimation_enabled FROM organizations WHERE name = $1
`

func (q *Queries) FindOrganizationByName(ctx context.Context, name organization.Name) (Organization, error) {
	row := q.db.QueryRow(ctx, findOrganizationByName, name)
	var i Organization
	err := row.Scan(
		&i.OrganizationID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.SessionRemember,
		&i.SessionTimeout,
		&i.Email,
		&i.CollaboratorAuthPolicy,
		&i.AllowForceDeleteWorkspaces,
		&i.CostEstimationEnabled,
	)
	return i, err
}

const findOrganizationByNameForUpdate = `-- name: FindOrganizationByNameForUpdate :one
SELECT organization_id, created_at, updated_at, name, session_remember, session_timeout, email, collaborator_auth_policy, allow_force_delete_workspaces, cost_estimation_enabled
FROM organizations
WHERE name = $1
FOR UPDATE
`

func (q *Queries) FindOrganizationByNameForUpdate(ctx context.Context, name organization.Name) (Organization, error) {
	row := q.db.QueryRow(ctx, findOrganizationByNameForUpdate, name)
	var i Organization
	err := row.Scan(
		&i.OrganizationID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.SessionRemember,
		&i.SessionTimeout,
		&i.Email,
		&i.CollaboratorAuthPolicy,
		&i.AllowForceDeleteWorkspaces,
		&i.CostEstimationEnabled,
	)
	return i, err
}

const findOrganizationNameByWorkspaceID = `-- name: FindOrganizationNameByWorkspaceID :one
SELECT organization_name
FROM workspaces
WHERE workspace_id = $1
`

func (q *Queries) FindOrganizationNameByWorkspaceID(ctx context.Context, workspaceID resource.ID) (organization.Name, error) {
	row := q.db.QueryRow(ctx, findOrganizationNameByWorkspaceID, workspaceID)
	var organization_name organization.Name
	err := row.Scan(&organization_name)
	return organization_name, err
}

const findOrganizations = `-- name: FindOrganizations :many
SELECT organization_id, created_at, updated_at, name, session_remember, session_timeout, email, collaborator_auth_policy, allow_force_delete_workspaces, cost_estimation_enabled
FROM organizations
WHERE name LIKE ANY($1::text[])
ORDER BY updated_at DESC
LIMIT $3::int OFFSET $2::int
`

type FindOrganizationsParams struct {
	Names  []pgtype.Text
	Offset pgtype.Int4
	Limit  pgtype.Int4
}

func (q *Queries) FindOrganizations(ctx context.Context, arg FindOrganizationsParams) ([]Organization, error) {
	rows, err := q.db.Query(ctx, findOrganizations, arg.Names, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Organization
	for rows.Next() {
		var i Organization
		if err := rows.Scan(
			&i.OrganizationID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.SessionRemember,
			&i.SessionTimeout,
			&i.Email,
			&i.CollaboratorAuthPolicy,
			&i.AllowForceDeleteWorkspaces,
			&i.CostEstimationEnabled,
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

const insertOrganization = `-- name: InsertOrganization :exec
INSERT INTO organizations (
    organization_id,
    created_at,
    updated_at,
    name,
    email,
    collaborator_auth_policy,
    cost_estimation_enabled,
    session_remember,
    session_timeout,
    allow_force_delete_workspaces
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
)
`

type InsertOrganizationParams struct {
	ID                         resource.ID
	CreatedAt                  pgtype.Timestamptz
	UpdatedAt                  pgtype.Timestamptz
	Name                       organization.Name
	Email                      pgtype.Text
	CollaboratorAuthPolicy     pgtype.Text
	CostEstimationEnabled      pgtype.Bool
	SessionRemember            pgtype.Int4
	SessionTimeout             pgtype.Int4
	AllowForceDeleteWorkspaces pgtype.Bool
}

func (q *Queries) InsertOrganization(ctx context.Context, arg InsertOrganizationParams) error {
	_, err := q.db.Exec(ctx, insertOrganization,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Name,
		arg.Email,
		arg.CollaboratorAuthPolicy,
		arg.CostEstimationEnabled,
		arg.SessionRemember,
		arg.SessionTimeout,
		arg.AllowForceDeleteWorkspaces,
	)
	return err
}

const updateOrganizationByName = `-- name: UpdateOrganizationByName :one
UPDATE organizations
SET
    name = $1,
    email = $2,
    collaborator_auth_policy = $3,
    cost_estimation_enabled = $4,
    session_remember = $5,
    session_timeout = $6,
    allow_force_delete_workspaces = $7,
    updated_at = $8
WHERE name = $9
RETURNING organization_id
`

type UpdateOrganizationByNameParams struct {
	NewName                    organization.Name
	Email                      pgtype.Text
	CollaboratorAuthPolicy     pgtype.Text
	CostEstimationEnabled      pgtype.Bool
	SessionRemember            pgtype.Int4
	SessionTimeout             pgtype.Int4
	AllowForceDeleteWorkspaces pgtype.Bool
	UpdatedAt                  pgtype.Timestamptz
	Name                       organization.Name
}

func (q *Queries) UpdateOrganizationByName(ctx context.Context, arg UpdateOrganizationByNameParams) (resource.ID, error) {
	row := q.db.QueryRow(ctx, updateOrganizationByName,
		arg.NewName,
		arg.Email,
		arg.CollaboratorAuthPolicy,
		arg.CostEstimationEnabled,
		arg.SessionRemember,
		arg.SessionTimeout,
		arg.AllowForceDeleteWorkspaces,
		arg.UpdatedAt,
		arg.Name,
	)
	var organization_id resource.ID
	err := row.Scan(&organization_id)
	return organization_id, err
}
