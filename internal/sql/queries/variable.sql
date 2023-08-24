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
    pggen.arg('variable_id'),
    pggen.arg('key'),
    pggen.arg('value'),
    pggen.arg('description'),
    pggen.arg('category'),
    pggen.arg('sensitive'),
    pggen.arg('hcl'),
    pggen.arg('version_id')
);

-- name: FindVariable :one
SELECT *
FROM variables
WHERE variable_id = pggen.arg('variable_id')
;

-- name: UpdateVariableByID :one
UPDATE variables
SET
    key = pggen.arg('key'),
    value = pggen.arg('value'),
    description = pggen.arg('description'),
    category = pggen.arg('category'),
    sensitive = pggen.arg('sensitive'),
    version_id = pggen.arg('version_id'),
    hcl = pggen.arg('hcl')
WHERE variable_id = pggen.arg('variable_id')
RETURNING variable_id
;

-- name: DeleteVariableByID :one
DELETE
FROM variables
WHERE variable_id = pggen.arg('variable_id')
RETURNING *
;
