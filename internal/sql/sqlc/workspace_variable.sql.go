// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: workspace_variable.sql

package sqlc

import (
	"context"

	"github.com/leg100/otf/internal/resource"
)

const deleteWorkspaceVariableByID = `-- name: DeleteWorkspaceVariableByID :one
DELETE
FROM workspace_variables wv USING variables v
WHERE wv.variable_id = $1
RETURNING wv.workspace_id, (v.*)::"variables" AS variable
`

type DeleteWorkspaceVariableByIDRow struct {
	WorkspaceID resource.ID
	Variable    Variable
}

func (q *Queries) DeleteWorkspaceVariableByID(ctx context.Context, variableID resource.ID) (DeleteWorkspaceVariableByIDRow, error) {
	row := q.db.QueryRow(ctx, deleteWorkspaceVariableByID, variableID)
	var i DeleteWorkspaceVariableByIDRow
	err := row.Scan(&i.WorkspaceID, &i.Variable)
	return i, err
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
	WorkspaceID resource.ID
	Variable    Variable
}

func (q *Queries) FindWorkspaceVariableByVariableID(ctx context.Context, variableID resource.ID) (FindWorkspaceVariableByVariableIDRow, error) {
	row := q.db.QueryRow(ctx, findWorkspaceVariableByVariableID, variableID)
	var i FindWorkspaceVariableByVariableIDRow
	err := row.Scan(&i.WorkspaceID, &i.Variable)
	return i, err
}

const findWorkspaceVariablesByWorkspaceID = `-- name: FindWorkspaceVariablesByWorkspaceID :many
SELECT v.variable_id, v.key, v.value, v.description, v.category, v.sensitive, v.hcl, v.version_id
FROM workspace_variables
JOIN variables v USING (variable_id)
WHERE workspace_id = $1
`

func (q *Queries) FindWorkspaceVariablesByWorkspaceID(ctx context.Context, workspaceID resource.ID) ([]Variable, error) {
	rows, err := q.db.Query(ctx, findWorkspaceVariablesByWorkspaceID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Variable
	for rows.Next() {
		var i Variable
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

func (q *Queries) InsertWorkspaceVariable(ctx context.Context, arg InsertWorkspaceVariableParams) error {
	_, err := q.db.Exec(ctx, insertWorkspaceVariable, arg.VariableID, arg.WorkspaceID)
	return err
}
