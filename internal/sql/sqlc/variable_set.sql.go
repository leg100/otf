// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: variable_set.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

const deleteVariableSetByID = `-- name: DeleteVariableSetByID :one
DELETE
FROM variable_sets
WHERE variable_set_id = $1
RETURNING variable_set_id, global, name, description, organization_name
`

func (q *Queries) DeleteVariableSetByID(ctx context.Context, variableSetID resource.ID) (VariableSet, error) {
	row := q.db.QueryRow(ctx, deleteVariableSetByID, variableSetID)
	var i VariableSet
	err := row.Scan(
		&i.VariableSetID,
		&i.Global,
		&i.Name,
		&i.Description,
		&i.OrganizationName,
	)
	return i, err
}

const deleteVariableSetVariable = `-- name: DeleteVariableSetVariable :one
DELETE
FROM variable_set_variables
WHERE variable_set_id = $1
AND variable_id = $2
RETURNING variable_set_id, variable_id
`

type DeleteVariableSetVariableParams struct {
	VariableSetID resource.ID
	VariableID    resource.ID
}

func (q *Queries) DeleteVariableSetVariable(ctx context.Context, arg DeleteVariableSetVariableParams) (VariableSetVariable, error) {
	row := q.db.QueryRow(ctx, deleteVariableSetVariable, arg.VariableSetID, arg.VariableID)
	var i VariableSetVariable
	err := row.Scan(&i.VariableSetID, &i.VariableID)
	return i, err
}

const deleteVariableSetWorkspace = `-- name: DeleteVariableSetWorkspace :one
DELETE
FROM variable_set_workspaces
WHERE variable_set_id = $1
AND workspace_id = $2
RETURNING variable_set_id, workspace_id
`

type DeleteVariableSetWorkspaceParams struct {
	VariableSetID resource.ID
	WorkspaceID   resource.ID
}

func (q *Queries) DeleteVariableSetWorkspace(ctx context.Context, arg DeleteVariableSetWorkspaceParams) (VariableSetWorkspace, error) {
	row := q.db.QueryRow(ctx, deleteVariableSetWorkspace, arg.VariableSetID, arg.WorkspaceID)
	var i VariableSetWorkspace
	err := row.Scan(&i.VariableSetID, &i.WorkspaceID)
	return i, err
}

const deleteVariableSetWorkspaces = `-- name: DeleteVariableSetWorkspaces :exec
DELETE
FROM variable_set_workspaces
WHERE variable_set_id = $1
`

func (q *Queries) DeleteVariableSetWorkspaces(ctx context.Context, variableSetID resource.ID) error {
	_, err := q.db.Exec(ctx, deleteVariableSetWorkspaces, variableSetID)
	return err
}

const findVariableSetBySetID = `-- name: FindVariableSetBySetID :one
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE vs.variable_set_id = $1
`

type FindVariableSetBySetIDRow struct {
	VariableSetID    resource.ID
	Global           pgtype.Bool
	Name             pgtype.Text
	Description      pgtype.Text
	OrganizationName pgtype.Text
	Variables        []Variable
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetBySetID(ctx context.Context, variableSetID resource.ID) (FindVariableSetBySetIDRow, error) {
	row := q.db.QueryRow(ctx, findVariableSetBySetID, variableSetID)
	var i FindVariableSetBySetIDRow
	err := row.Scan(
		&i.VariableSetID,
		&i.Global,
		&i.Name,
		&i.Description,
		&i.OrganizationName,
		&i.Variables,
		&i.WorkspaceIds,
	)
	return i, err
}

const findVariableSetByVariableID = `-- name: FindVariableSetByVariableID :one
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN variable_set_variables vsv USING (variable_set_id)
WHERE vsv.variable_id = $1
`

type FindVariableSetByVariableIDRow struct {
	VariableSetID    resource.ID
	Global           pgtype.Bool
	Name             pgtype.Text
	Description      pgtype.Text
	OrganizationName pgtype.Text
	Variables        []Variable
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetByVariableID(ctx context.Context, variableID resource.ID) (FindVariableSetByVariableIDRow, error) {
	row := q.db.QueryRow(ctx, findVariableSetByVariableID, variableID)
	var i FindVariableSetByVariableIDRow
	err := row.Scan(
		&i.VariableSetID,
		&i.Global,
		&i.Name,
		&i.Description,
		&i.OrganizationName,
		&i.Variables,
		&i.WorkspaceIds,
	)
	return i, err
}

const findVariableSetForUpdate = `-- name: FindVariableSetForUpdate :one
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE vs.variable_set_id = $1
FOR UPDATE OF vs
`

type FindVariableSetForUpdateRow struct {
	VariableSetID    resource.ID
	Global           pgtype.Bool
	Name             pgtype.Text
	Description      pgtype.Text
	OrganizationName pgtype.Text
	Variables        []Variable
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetForUpdate(ctx context.Context, variableSetID resource.ID) (FindVariableSetForUpdateRow, error) {
	row := q.db.QueryRow(ctx, findVariableSetForUpdate, variableSetID)
	var i FindVariableSetForUpdateRow
	err := row.Scan(
		&i.VariableSetID,
		&i.Global,
		&i.Name,
		&i.Description,
		&i.OrganizationName,
		&i.Variables,
		&i.WorkspaceIds,
	)
	return i, err
}

const findVariableSetsByOrganization = `-- name: FindVariableSetsByOrganization :many
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE organization_name = $1
`

type FindVariableSetsByOrganizationRow struct {
	VariableSetID    resource.ID
	Global           pgtype.Bool
	Name             pgtype.Text
	Description      pgtype.Text
	OrganizationName pgtype.Text
	Variables        []Variable
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetsByOrganization(ctx context.Context, organizationName pgtype.Text) ([]FindVariableSetsByOrganizationRow, error) {
	rows, err := q.db.Query(ctx, findVariableSetsByOrganization, organizationName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindVariableSetsByOrganizationRow
	for rows.Next() {
		var i FindVariableSetsByOrganizationRow
		if err := rows.Scan(
			&i.VariableSetID,
			&i.Global,
			&i.Name,
			&i.Description,
			&i.OrganizationName,
			&i.Variables,
			&i.WorkspaceIds,
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

const findVariableSetsByWorkspace = `-- name: FindVariableSetsByWorkspace :many
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN variable_set_workspaces vsw USING (variable_set_id)
WHERE vsw.workspace_id = $1
UNION
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN (organizations o JOIN workspaces w ON o.name = w.organization_name) ON o.name = vs.organization_name
WHERE vs.global IS true
AND w.workspace_id = $1
`

type FindVariableSetsByWorkspaceRow struct {
	VariableSetID    resource.ID
	Global           pgtype.Bool
	Name             pgtype.Text
	Description      pgtype.Text
	OrganizationName pgtype.Text
	Variables        []Variable
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetsByWorkspace(ctx context.Context, workspaceID resource.ID) ([]FindVariableSetsByWorkspaceRow, error) {
	rows, err := q.db.Query(ctx, findVariableSetsByWorkspace, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindVariableSetsByWorkspaceRow
	for rows.Next() {
		var i FindVariableSetsByWorkspaceRow
		if err := rows.Scan(
			&i.VariableSetID,
			&i.Global,
			&i.Name,
			&i.Description,
			&i.OrganizationName,
			&i.Variables,
			&i.WorkspaceIds,
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

const insertVariableSet = `-- name: InsertVariableSet :exec
INSERT INTO variable_sets (
    variable_set_id,
    global,
    name,
    description,
    organization_name
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
`

type InsertVariableSetParams struct {
	VariableSetID    resource.ID
	Global           pgtype.Bool
	Name             pgtype.Text
	Description      pgtype.Text
	OrganizationName pgtype.Text
}

func (q *Queries) InsertVariableSet(ctx context.Context, arg InsertVariableSetParams) error {
	_, err := q.db.Exec(ctx, insertVariableSet,
		arg.VariableSetID,
		arg.Global,
		arg.Name,
		arg.Description,
		arg.OrganizationName,
	)
	return err
}

const insertVariableSetVariable = `-- name: InsertVariableSetVariable :exec
INSERT INTO variable_set_variables (
    variable_set_id,
    variable_id
) VALUES (
    $1,
    $2
)
`

type InsertVariableSetVariableParams struct {
	VariableSetID resource.ID
	VariableID    resource.ID
}

func (q *Queries) InsertVariableSetVariable(ctx context.Context, arg InsertVariableSetVariableParams) error {
	_, err := q.db.Exec(ctx, insertVariableSetVariable, arg.VariableSetID, arg.VariableID)
	return err
}

const insertVariableSetWorkspace = `-- name: InsertVariableSetWorkspace :exec
INSERT INTO variable_set_workspaces (
    variable_set_id,
    workspace_id
) VALUES (
    $1,
    $2
)
`

type InsertVariableSetWorkspaceParams struct {
	VariableSetID resource.ID
	WorkspaceID   resource.ID
}

func (q *Queries) InsertVariableSetWorkspace(ctx context.Context, arg InsertVariableSetWorkspaceParams) error {
	_, err := q.db.Exec(ctx, insertVariableSetWorkspace, arg.VariableSetID, arg.WorkspaceID)
	return err
}

const updateVariableSetByID = `-- name: UpdateVariableSetByID :one
UPDATE variable_sets
SET
    global = $1,
    name = $2,
    description = $3
WHERE variable_set_id = $4
RETURNING variable_set_id
`

type UpdateVariableSetByIDParams struct {
	Global        pgtype.Bool
	Name          pgtype.Text
	Description   pgtype.Text
	VariableSetID resource.ID
}

func (q *Queries) UpdateVariableSetByID(ctx context.Context, arg UpdateVariableSetByIDParams) (resource.ID, error) {
	row := q.db.QueryRow(ctx, updateVariableSetByID,
		arg.Global,
		arg.Name,
		arg.Description,
		arg.VariableSetID,
	)
	var variable_set_id resource.ID
	err := row.Scan(&variable_set_id)
	return variable_set_id, err
}
