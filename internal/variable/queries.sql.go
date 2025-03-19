// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queries.sql

package variable

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

const deleteVariableByID = `-- name: DeleteVariableByID :one
DELETE
FROM variables
WHERE variable_id = $1
RETURNING variable_id, key, value, description, category, sensitive, hcl, version_id
`

func (q *Queries) DeleteVariableByID(ctx context.Context, db DBTX, variableID resource.ID) (VariableModel, error) {
	row := db.QueryRow(ctx, deleteVariableByID, variableID)
	var i VariableModel
	err := row.Scan(
		&i.VariableID,
		&i.Key,
		&i.Value,
		&i.Description,
		&i.Category,
		&i.Sensitive,
		&i.HCL,
		&i.VersionID,
	)
	return i, err
}

const deleteVariableSetByID = `-- name: DeleteVariableSetByID :one
DELETE
FROM variable_sets
WHERE variable_set_id = $1
RETURNING variable_set_id, global, name, description, organization_name
`

func (q *Queries) DeleteVariableSetByID(ctx context.Context, db DBTX, variableSetID resource.ID) (VariableSetModel, error) {
	row := db.QueryRow(ctx, deleteVariableSetByID, variableSetID)
	var i VariableSetModel
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

func (q *Queries) DeleteVariableSetVariable(ctx context.Context, db DBTX, arg DeleteVariableSetVariableParams) (VariableSetVariable, error) {
	row := db.QueryRow(ctx, deleteVariableSetVariable, arg.VariableSetID, arg.VariableID)
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

func (q *Queries) DeleteVariableSetWorkspace(ctx context.Context, db DBTX, arg DeleteVariableSetWorkspaceParams) (VariableSetWorkspace, error) {
	row := db.QueryRow(ctx, deleteVariableSetWorkspace, arg.VariableSetID, arg.WorkspaceID)
	var i VariableSetWorkspace
	err := row.Scan(&i.VariableSetID, &i.WorkspaceID)
	return i, err
}

const deleteVariableSetWorkspaces = `-- name: DeleteVariableSetWorkspaces :exec
DELETE
FROM variable_set_workspaces
WHERE variable_set_id = $1
`

func (q *Queries) DeleteVariableSetWorkspaces(ctx context.Context, db DBTX, variableSetID resource.ID) error {
	_, err := db.Exec(ctx, deleteVariableSetWorkspaces, variableSetID)
	return err
}

const deleteWorkspaceVariableByID = `-- name: DeleteWorkspaceVariableByID :one
DELETE
FROM workspace_variables wv USING variables v
WHERE wv.variable_id = $1
RETURNING wv.workspace_id, (v.*)::"variables" AS variable
`

type DeleteWorkspaceVariableByIDRow struct {
	WorkspaceID   resource.ID
	VariableModel VariableModel
}

func (q *Queries) DeleteWorkspaceVariableByID(ctx context.Context, db DBTX, variableID resource.ID) (DeleteWorkspaceVariableByIDRow, error) {
	row := db.QueryRow(ctx, deleteWorkspaceVariableByID, variableID)
	var i DeleteWorkspaceVariableByIDRow
	err := row.Scan(&i.WorkspaceID, &i.VariableModel)
	return i, err
}

const findVariable = `-- name: FindVariable :one
SELECT variable_id, key, value, description, category, sensitive, hcl, version_id
FROM variables
WHERE variable_id = $1
`

func (q *Queries) FindVariable(ctx context.Context, db DBTX, variableID resource.ID) (VariableModel, error) {
	row := db.QueryRow(ctx, findVariable, variableID)
	var i VariableModel
	err := row.Scan(
		&i.VariableID,
		&i.Key,
		&i.Value,
		&i.Description,
		&i.Category,
		&i.Sensitive,
		&i.HCL,
		&i.VersionID,
	)
	return i, err
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
	OrganizationName organization.Name
	Variables        []VariableModel
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetBySetID(ctx context.Context, db DBTX, variableSetID resource.ID) (FindVariableSetBySetIDRow, error) {
	row := db.QueryRow(ctx, findVariableSetBySetID, variableSetID)
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
	OrganizationName organization.Name
	Variables        []VariableModel
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetByVariableID(ctx context.Context, db DBTX, variableID resource.ID) (FindVariableSetByVariableIDRow, error) {
	row := db.QueryRow(ctx, findVariableSetByVariableID, variableID)
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
	OrganizationName organization.Name
	Variables        []VariableModel
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetForUpdate(ctx context.Context, db DBTX, variableSetID resource.ID) (FindVariableSetForUpdateRow, error) {
	row := db.QueryRow(ctx, findVariableSetForUpdate, variableSetID)
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
	OrganizationName organization.Name
	Variables        []VariableModel
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetsByOrganization(ctx context.Context, db DBTX, organizationName organization.Name) ([]FindVariableSetsByOrganizationRow, error) {
	rows, err := db.Query(ctx, findVariableSetsByOrganization, organizationName)
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
	OrganizationName organization.Name
	Variables        []VariableModel
	WorkspaceIds     []pgtype.Text
}

func (q *Queries) FindVariableSetsByWorkspace(ctx context.Context, db DBTX, workspaceID resource.ID) ([]FindVariableSetsByWorkspaceRow, error) {
	rows, err := db.Query(ctx, findVariableSetsByWorkspace, workspaceID)
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

const findWorkspaceVariableByVariableID = `-- name: FindWorkspaceVariableByVariableID :one
SELECT
    workspace_id,
    v::"variables" AS variable
FROM workspace_variables
JOIN variables v USING (variable_id)
WHERE v.variable_id = $1
`

type FindWorkspaceVariableByVariableIDRow struct {
	WorkspaceID   resource.ID
	VariableModel VariableModel
}

func (q *Queries) FindWorkspaceVariableByVariableID(ctx context.Context, db DBTX, variableID resource.ID) (FindWorkspaceVariableByVariableIDRow, error) {
	row := db.QueryRow(ctx, findWorkspaceVariableByVariableID, variableID)
	var i FindWorkspaceVariableByVariableIDRow
	err := row.Scan(&i.WorkspaceID, &i.VariableModel)
	return i, err
}

const findWorkspaceVariablesByWorkspaceID = `-- name: FindWorkspaceVariablesByWorkspaceID :many
SELECT v.variable_id, v.key, v.value, v.description, v.category, v.sensitive, v.hcl, v.version_id
FROM workspace_variables
JOIN variables v USING (variable_id)
WHERE workspace_id = $1
`

func (q *Queries) FindWorkspaceVariablesByWorkspaceID(ctx context.Context, db DBTX, workspaceID resource.ID) ([]VariableModel, error) {
	rows, err := db.Query(ctx, findWorkspaceVariablesByWorkspaceID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []VariableModel
	for rows.Next() {
		var i VariableModel
		if err := rows.Scan(
			&i.VariableID,
			&i.Key,
			&i.Value,
			&i.Description,
			&i.Category,
			&i.Sensitive,
			&i.HCL,
			&i.VersionID,
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

const insertVariable = `-- name: InsertVariable :exec
INSERT INTO variables (
    variable_id,
    key,
    value,
    description,
    category,
    sensitive,
    hcl,
    version_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
`

type InsertVariableParams struct {
	VariableID  resource.ID
	Key         pgtype.Text
	Value       pgtype.Text
	Description pgtype.Text
	Category    pgtype.Text
	Sensitive   pgtype.Bool
	HCL         pgtype.Bool
	VersionID   pgtype.Text
}

func (q *Queries) InsertVariable(ctx context.Context, db DBTX, arg InsertVariableParams) error {
	_, err := db.Exec(ctx, insertVariable,
		arg.VariableID,
		arg.Key,
		arg.Value,
		arg.Description,
		arg.Category,
		arg.Sensitive,
		arg.HCL,
		arg.VersionID,
	)
	return err
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
	OrganizationName organization.Name
}

// variable sets
func (q *Queries) InsertVariableSet(ctx context.Context, db DBTX, arg InsertVariableSetParams) error {
	_, err := db.Exec(ctx, insertVariableSet,
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

func (q *Queries) InsertVariableSetVariable(ctx context.Context, db DBTX, arg InsertVariableSetVariableParams) error {
	_, err := db.Exec(ctx, insertVariableSetVariable, arg.VariableSetID, arg.VariableID)
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

func (q *Queries) InsertVariableSetWorkspace(ctx context.Context, db DBTX, arg InsertVariableSetWorkspaceParams) error {
	_, err := db.Exec(ctx, insertVariableSetWorkspace, arg.VariableSetID, arg.WorkspaceID)
	return err
}

const insertWorkspaceVariable = `-- name: InsertWorkspaceVariable :exec

INSERT INTO workspace_variables (
    variable_id,
    workspace_id
) VALUES (
    $1,
    $2
)
`

type InsertWorkspaceVariableParams struct {
	VariableID  resource.ID
	WorkspaceID resource.ID
}

// workspace variables
func (q *Queries) InsertWorkspaceVariable(ctx context.Context, db DBTX, arg InsertWorkspaceVariableParams) error {
	_, err := db.Exec(ctx, insertWorkspaceVariable, arg.VariableID, arg.WorkspaceID)
	return err
}

const updateVariableByID = `-- name: UpdateVariableByID :one
UPDATE variables
SET
    key = $1,
    value = $2,
    description = $3,
    category = $4,
    sensitive = $5,
    version_id = $6,
    hcl = $7
WHERE variable_id = $8
RETURNING variable_id
`

type UpdateVariableByIDParams struct {
	Key         pgtype.Text
	Value       pgtype.Text
	Description pgtype.Text
	Category    pgtype.Text
	Sensitive   pgtype.Bool
	VersionID   pgtype.Text
	HCL         pgtype.Bool
	VariableID  resource.ID
}

func (q *Queries) UpdateVariableByID(ctx context.Context, db DBTX, arg UpdateVariableByIDParams) (resource.ID, error) {
	row := db.QueryRow(ctx, updateVariableByID,
		arg.Key,
		arg.Value,
		arg.Description,
		arg.Category,
		arg.Sensitive,
		arg.VersionID,
		arg.HCL,
		arg.VariableID,
	)
	var variable_id resource.ID
	err := row.Scan(&variable_id)
	return variable_id, err
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

func (q *Queries) UpdateVariableSetByID(ctx context.Context, db DBTX, arg UpdateVariableSetByIDParams) (resource.ID, error) {
	row := db.QueryRow(ctx, updateVariableSetByID,
		arg.Global,
		arg.Name,
		arg.Description,
		arg.VariableSetID,
	)
	var variable_set_id resource.ID
	err := row.Scan(&variable_set_id)
	return variable_set_id, err
}
