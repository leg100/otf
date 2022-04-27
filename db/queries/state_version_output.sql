-- InsertStateVersionOutput inserts a state_version_output and returns the entire row.
--
-- name: InsertStateVersionOutput :one
INSERT INTO state_version_outputs (
    state_version_output_id,
    created_at,
    updated_at,
    name,
    sensitive,
    type,
    value,
    state_version_id
) VALUES (
    pggen.arg('ID'),
    NOW(),
    NOW(),
    pggen.arg('Name'),
    pggen.arg('Sensitive'),
    pggen.arg('Type'),
    pggen.arg('Value'),
    pggen.arg('StateVersionID')
)
RETURNING *;
