-- name: InsertStateVersionOutput :exec
INSERT INTO state_version_outputs (
    state_version_output_id,
    name,
    sensitive,
    type,
    value,
    state_version_id
) VALUES (
    pggen.arg('id'),
    pggen.arg('name'),
    pggen.arg('sensitive'),
    pggen.arg('type'),
    pggen.arg('value'),
    pggen.arg('state_version_id')
);
