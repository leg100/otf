-- name: InsertVariable :exec
INSERT INTO variables (
    variable_id,
    key,
    value,
    description,
    category,
    sensitive,
    hcl,
    workspace_id
) VALUES (
    pggen.arg('variable_id'),
    pggen.arg('key'),
    pggen.arg('value'),
    pggen.arg('description'),
    pggen.arg('category'),
    pggen.arg('sensitive'),
    pggen.arg('hcl'),
    pggen.arg('workspace_id')
);

-- name: FindVariables :many
SELECT *
FROM variables
WHERE workspace_id = pggen.arg('workspace_id')
;

-- name: FindVariable :one
SELECT *
FROM variables
WHERE variable_id = pggen.arg('variable_id')
;

-- name: FindVariableForUpdate :one
SELECT *
FROM variables
WHERE variable_id = pggen.arg('variable_id')
FOR UPDATE;

-- name: UpdateVariable :one
UPDATE variables
SET
    key = pggen.arg('key'),
    value = pggen.arg('value'),
    description = pggen.arg('description'),
    category = pggen.arg('category'),
    sensitive = pggen.arg('sensitive'),
    hcl = pggen.arg('hcl')
WHERE variable_id = pggen.arg('variable_id')
RETURNING variable_id
;

-- name: DeleteVariable :one
DELETE
FROM variables
WHERE variable_id = pggen.arg('variable_id')
RETURNING *
;
