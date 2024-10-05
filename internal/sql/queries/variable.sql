-- name: InsertVariable :exec
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
    sqlc.arg('variable_id'),
    sqlc.arg('key'),
    sqlc.arg('value'),
    sqlc.arg('description'),
    sqlc.arg('category'),
    sqlc.arg('sensitive'),
    sqlc.arg('hcl'),
    sqlc.arg('version_id')
);

-- name: FindVariable :one
SELECT *
FROM variables
WHERE variable_id = sqlc.arg('variable_id')
;

-- name: UpdateVariableByID :one
UPDATE variables
SET
    key = sqlc.arg('key'),
    value = sqlc.arg('value'),
    description = sqlc.arg('description'),
    category = sqlc.arg('category'),
    sensitive = sqlc.arg('sensitive'),
    version_id = sqlc.arg('version_id'),
    hcl = sqlc.arg('hcl')
WHERE variable_id = sqlc.arg('variable_id')
RETURNING variable_id
;

-- name: DeleteVariableByID :one
DELETE
FROM variables
WHERE variable_id = sqlc.arg('variable_id')
RETURNING *
;
