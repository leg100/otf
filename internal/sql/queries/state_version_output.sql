-- name: InsertStateVersionOutput :exec
INSERT INTO state_version_outputs (
    state_version_output_id,
    name,
    sensitive,
    type,
    value,
    state_version_id
) VALUES (
    sqlc.arg('id'),
    sqlc.arg('name'),
    sqlc.arg('sensitive'),
    sqlc.arg('type'),
    sqlc.arg('value'),
    sqlc.arg('state_version_id')
);

-- name: FindStateVersionOutputByID :one
SELECT *
FROM state_version_outputs
WHERE state_version_output_id = sqlc.arg('id')
;
