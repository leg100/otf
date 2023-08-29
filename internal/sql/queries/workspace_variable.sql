-- name: InsertWorkspaceVariable :exec
INSERT INTO workspace_variables (
    variable_id,
    workspace_id
) VALUES (
    pggen.arg('variable_id'),
    pggen.arg('workspace_id')
);

-- name: FindVariablesByWorkspaceID :many
SELECT v.*
FROM workspace_variables
JOIN variables v USING (variable_id)
WHERE workspace_id = pggen.arg('workspace_id')
;

-- name: FindWorkspaceVariableByID :one
SELECT workspace_id, (v.*)::"variables" AS variable
FROM workspace_variables
JOIN variables v USING (variable_id)
WHERE variable_id = pggen.arg('variable_id')
;

-- name: DeleteWorkspaceVariableByID :one
DELETE
FROM workspace_variables
WHERE variable_id = pggen.arg('variable_id')
AND workspace_id = pggen.arg('workspace_id')
RETURNING *;
