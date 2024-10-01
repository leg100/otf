// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: team.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteTeamByID = `-- name: DeleteTeamByID :one
DELETE
FROM teams
WHERE team_id = $1
RETURNING team_id
`

func (q *Queries) DeleteTeamByID(ctx context.Context, teamID pgtype.Text) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, deleteTeamByID, teamID)
	var team_id pgtype.Text
	err := row.Scan(&team_id)
	return team_id, err
}

const findTeamByID = `-- name: FindTeamByID :one
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams
WHERE team_id = $1
`

func (q *Queries) FindTeamByID(ctx context.Context, teamID pgtype.Text) (Team, error) {
	row := q.db.QueryRow(ctx, findTeamByID, teamID)
	var i Team
	err := row.Scan(
		&i.TeamID,
		&i.Name,
		&i.CreatedAt,
		&i.PermissionManageWorkspaces,
		&i.PermissionManageVCS,
		&i.PermissionManageModules,
		&i.OrganizationName,
		&i.SSOTeamID,
		&i.Visibility,
		&i.PermissionManagePolicies,
		&i.PermissionManagePolicyOverrides,
		&i.PermissionManageProviders,
	)
	return i, err
}

const findTeamByIDForUpdate = `-- name: FindTeamByIDForUpdate :one
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams t
WHERE team_id = $1
FOR UPDATE OF t
`

func (q *Queries) FindTeamByIDForUpdate(ctx context.Context, teamID pgtype.Text) (Team, error) {
	row := q.db.QueryRow(ctx, findTeamByIDForUpdate, teamID)
	var i Team
	err := row.Scan(
		&i.TeamID,
		&i.Name,
		&i.CreatedAt,
		&i.PermissionManageWorkspaces,
		&i.PermissionManageVCS,
		&i.PermissionManageModules,
		&i.OrganizationName,
		&i.SSOTeamID,
		&i.Visibility,
		&i.PermissionManagePolicies,
		&i.PermissionManagePolicyOverrides,
		&i.PermissionManageProviders,
	)
	return i, err
}

const findTeamByName = `-- name: FindTeamByName :one
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams
WHERE name              = $1
AND   organization_name = $2
`

type FindTeamByNameParams struct {
	Name             pgtype.Text
	OrganizationName pgtype.Text
}

func (q *Queries) FindTeamByName(ctx context.Context, arg FindTeamByNameParams) (Team, error) {
	row := q.db.QueryRow(ctx, findTeamByName, arg.Name, arg.OrganizationName)
	var i Team
	err := row.Scan(
		&i.TeamID,
		&i.Name,
		&i.CreatedAt,
		&i.PermissionManageWorkspaces,
		&i.PermissionManageVCS,
		&i.PermissionManageModules,
		&i.OrganizationName,
		&i.SSOTeamID,
		&i.Visibility,
		&i.PermissionManagePolicies,
		&i.PermissionManagePolicyOverrides,
		&i.PermissionManageProviders,
	)
	return i, err
}

const findTeamByTokenID = `-- name: FindTeamByTokenID :one
SELECT t.team_id, t.name, t.created_at, t.permission_manage_workspaces, t.permission_manage_vcs, t.permission_manage_modules, t.organization_name, t.sso_team_id, t.visibility, t.permission_manage_policies, t.permission_manage_policy_overrides, t.permission_manage_providers
FROM teams t
JOIN team_tokens tt USING (team_id)
WHERE tt.team_token_id = $1
`

func (q *Queries) FindTeamByTokenID(ctx context.Context, tokenID pgtype.Text) (Team, error) {
	row := q.db.QueryRow(ctx, findTeamByTokenID, tokenID)
	var i Team
	err := row.Scan(
		&i.TeamID,
		&i.Name,
		&i.CreatedAt,
		&i.PermissionManageWorkspaces,
		&i.PermissionManageVCS,
		&i.PermissionManageModules,
		&i.OrganizationName,
		&i.SSOTeamID,
		&i.Visibility,
		&i.PermissionManagePolicies,
		&i.PermissionManagePolicyOverrides,
		&i.PermissionManageProviders,
	)
	return i, err
}

const findTeamsByOrg = `-- name: FindTeamsByOrg :many
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams
WHERE organization_name = $1
`

func (q *Queries) FindTeamsByOrg(ctx context.Context, organizationName pgtype.Text) ([]Team, error) {
	rows, err := q.db.Query(ctx, findTeamsByOrg, organizationName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Team
	for rows.Next() {
		var i Team
		if err := rows.Scan(
			&i.TeamID,
			&i.Name,
			&i.CreatedAt,
			&i.PermissionManageWorkspaces,
			&i.PermissionManageVCS,
			&i.PermissionManageModules,
			&i.OrganizationName,
			&i.SSOTeamID,
			&i.Visibility,
			&i.PermissionManagePolicies,
			&i.PermissionManagePolicyOverrides,
			&i.PermissionManageProviders,
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

const insertTeam = `-- name: InsertTeam :exec
INSERT INTO teams (
    team_id,
    name,
    created_at,
    organization_name,
    visibility,
    sso_team_id,
    permission_manage_workspaces,
    permission_manage_vcs,
    permission_manage_modules,
    permission_manage_providers,
    permission_manage_policies,
    permission_manage_policy_overrides
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
    $10,
    $11,
    $12
)
`

type InsertTeamParams struct {
	ID                              pgtype.Text
	Name                            pgtype.Text
	CreatedAt                       pgtype.Timestamptz
	OrganizationName                pgtype.Text
	Visibility                      pgtype.Text
	SSOTeamID                       pgtype.Text
	PermissionManageWorkspaces      pgtype.Bool
	PermissionManageVCS             pgtype.Bool
	PermissionManageModules         pgtype.Bool
	PermissionManageProviders       pgtype.Bool
	PermissionManagePolicies        pgtype.Bool
	PermissionManagePolicyOverrides pgtype.Bool
}

func (q *Queries) InsertTeam(ctx context.Context, arg InsertTeamParams) error {
	_, err := q.db.Exec(ctx, insertTeam,
		arg.ID,
		arg.Name,
		arg.CreatedAt,
		arg.OrganizationName,
		arg.Visibility,
		arg.SSOTeamID,
		arg.PermissionManageWorkspaces,
		arg.PermissionManageVCS,
		arg.PermissionManageModules,
		arg.PermissionManageProviders,
		arg.PermissionManagePolicies,
		arg.PermissionManagePolicyOverrides,
	)
	return err
}

const updateTeamByID = `-- name: UpdateTeamByID :one
UPDATE teams
SET
    name = $1,
    visibility = $2,
    sso_team_id = $3,
    permission_manage_workspaces = $4,
    permission_manage_vcs = $5,
    permission_manage_modules = $6,
    permission_manage_providers = $7,
    permission_manage_policies = $8,
    permission_manage_policy_overrides = $9
WHERE team_id = $10
RETURNING team_id
`

type UpdateTeamByIDParams struct {
	Name                            pgtype.Text
	Visibility                      pgtype.Text
	SSOTeamID                       pgtype.Text
	PermissionManageWorkspaces      pgtype.Bool
	PermissionManageVCS             pgtype.Bool
	PermissionManageModules         pgtype.Bool
	PermissionManageProviders       pgtype.Bool
	PermissionManagePolicies        pgtype.Bool
	PermissionManagePolicyOverrides pgtype.Bool
	TeamID                          pgtype.Text
}

func (q *Queries) UpdateTeamByID(ctx context.Context, arg UpdateTeamByIDParams) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, updateTeamByID,
		arg.Name,
		arg.Visibility,
		arg.SSOTeamID,
		arg.PermissionManageWorkspaces,
		arg.PermissionManageVCS,
		arg.PermissionManageModules,
		arg.PermissionManageProviders,
		arg.PermissionManagePolicies,
		arg.PermissionManagePolicyOverrides,
		arg.TeamID,
	)
	var team_id pgtype.Text
	err := row.Scan(&team_id)
	return team_id, err
}
