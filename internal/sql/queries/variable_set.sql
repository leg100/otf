-- name: InsertVariableSet :exec
INSERT INTO variable_sets (
    variable_set_id,
    global,
    name,
    description,
    organization_name
) VALUES (
    pggen.arg('variable_set_id'),
    pggen.arg('global'),
    pggen.arg('name'),
    pggen.arg('description'),
    pggen.arg('organization_name')
);

-- name: FindVariableSetsByOrganization :many
SELECT
    *,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE organization_name = pggen.arg('organization_name');

-- name: FindVariableSetsByWorkspace :many
SELECT
    vs.*,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN variable_set_workspaces vsw USING (variable_set_id)
WHERE workspace_id = pggen.arg('workspace_id')
UNION
SELECT
    vs.*,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN (organizations o JOIN workspaces w ON o.name = w.organization_name) ON o.name = vs.organization_name
WHERE vs.global IS true
AND w.workspace_id = pggen.arg('workspace_id');

-- name: FindVariableSetBySetID :one
SELECT
    *,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE vs.variable_set_id = pggen.arg('variable_set_id');

-- name: FindVariableSetByVariableID :one
SELECT
    vs.*,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN variable_set_variables vsv USING (variable_set_id)
WHERE vsv.variable_id = pggen.arg('variable_id');

-- name: FindVariableSetForUpdate :one
SELECT
    *,
    (
        SELECT array_agg(v.*) AS variables
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id) AS workspace_ids
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
        GROUP BY variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE variable_set_id = pggen.arg('variable_set_id')
FOR UPDATE OF vs;

-- name: UpdateVariableSetByID :one
UPDATE variable_sets
SET
    global = pggen.arg('global'),
    name = pggen.arg('name'),
    description = pggen.arg('description')
WHERE variable_set_id = pggen.arg('variable_set_id')
RETURNING variable_set_id;

-- name: DeleteVariableSetByID :one
DELETE
FROM variable_sets
WHERE variable_set_id = pggen.arg('variable_set_id')
RETURNING *;

-- name: InsertVariableSetVariable :exec
INSERT INTO variable_set_variables (
    variable_set_id,
    variable_id
) VALUES (
    pggen.arg('variable_set_id'),
    pggen.arg('variable_id')
);

-- name: DeleteVariableSetVariable :one
DELETE
FROM variable_set_variables
WHERE variable_set_id = pggen.arg('variable_set_id')
AND variable_id = pggen.arg('variable_id')
RETURNING *;

-- name: InsertVariableSetWorkspace :exec
INSERT INTO variable_set_workspaces (
    variable_set_id,
    workspace_id
) VALUES (
    pggen.arg('variable_set_id'),
    pggen.arg('workspace_id')
);

-- name: DeleteVariableSetWorkspace :one
DELETE
FROM variable_set_workspaces
WHERE variable_set_id = pggen.arg('variable_set_id')
AND workspace_id = pggen.arg('workspace_id')
RETURNING *;

-- name: DeleteVariableSetWorkspaces :exec
DELETE
FROM variable_set_workspaces
WHERE variable_set_id = pggen.arg('variable_set_id');
