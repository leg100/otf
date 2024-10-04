-- name: InsertVariableSet :exec
INSERT INTO variable_sets (
    variable_set_id,
    global,
    name,
    description,
    organization_name
) VALUES (
    sqlc.arg('variable_set_id'),
    sqlc.arg('global'),
    sqlc.arg('name'),
    sqlc.arg('description'),
    sqlc.arg('organization_name')
);

-- name: FindVariableSetsByOrganization :many
SELECT
    vs.*,
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
WHERE organization_name = sqlc.arg('organization_name')
;

-- name: FindVariableSetsByWorkspace :many
SELECT
    vs.*,
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
WHERE vsw.workspace_id = sqlc.arg('workspace_id')
UNION
SELECT
    vs.*,
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
AND w.workspace_id = sqlc.arg('workspace_id')
;

-- name: FindVariableSetBySetID :one
SELECT
    vs.*,
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
WHERE vs.variable_set_id = sqlc.arg('variable_set_id')
;

-- name: FindVariableSetByVariableID :one
SELECT
    vs.*,
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
WHERE vsv.variable_id = sqlc.arg('variable_id')
;

-- name: FindVariableSetForUpdate :one
SELECT
    vs.*,
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
WHERE vs.variable_set_id = sqlc.arg('variable_set_id')
FOR UPDATE OF vs;

-- name: UpdateVariableSetByID :one
UPDATE variable_sets
SET
    global = sqlc.arg('global'),
    name = sqlc.arg('name'),
    description = sqlc.arg('description')
WHERE variable_set_id = sqlc.arg('variable_set_id')
RETURNING variable_set_id;

-- name: DeleteVariableSetByID :one
DELETE
FROM variable_sets
WHERE variable_set_id = sqlc.arg('variable_set_id')
RETURNING *;

-- name: InsertVariableSetVariable :exec
INSERT INTO variable_set_variables (
    variable_set_id,
    variable_id
) VALUES (
    sqlc.arg('variable_set_id'),
    sqlc.arg('variable_id')
);

-- name: DeleteVariableSetVariable :one
DELETE
FROM variable_set_variables
WHERE variable_set_id = sqlc.arg('variable_set_id')
AND variable_id = sqlc.arg('variable_id')
RETURNING *;

-- name: InsertVariableSetWorkspace :exec
INSERT INTO variable_set_workspaces (
    variable_set_id,
    workspace_id
) VALUES (
    sqlc.arg('variable_set_id'),
    sqlc.arg('workspace_id')
);

-- name: DeleteVariableSetWorkspace :one
DELETE
FROM variable_set_workspaces
WHERE variable_set_id = sqlc.arg('variable_set_id')
AND workspace_id = sqlc.arg('workspace_id')
RETURNING *;

-- name: DeleteVariableSetWorkspaces :exec
DELETE
FROM variable_set_workspaces
WHERE variable_set_id = sqlc.arg('variable_set_id');
