-- name: InsertWorkspaceVariable :exec
INSERT INTO workspace_variables (
    variable_id,
    workspace_id
) VALUES (
    sqlc.arg('variable_id'),
    sqlc.arg('workspace_id')
);

-- name: FindWorkspaceVariablesByWorkspaceID :many
SELECT v.*
FROM workspace_variables
JOIN variables v USING (variable_id)
WHERE workspace_id = sqlc.arg('workspace_id');

-- name: FindWorkspaceVariableByVariableID :one
SELECT
    workspace_id,
    v::"variables" AS variable
FROM workspace_variables
JOIN variables v USING (variable_id)
WHERE v.variable_id = sqlc.arg('variable_id');

-- name: DeleteWorkspaceVariableByID :one
DELETE
FROM workspace_variables wv USING variables v
WHERE wv.variable_id = sqlc.arg('variable_id')
RETURNING wv.workspace_id, (v.*)::"variables" AS variable;
