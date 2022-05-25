-- name: InsertStateVersionOutput :exec
INSERT INTO state_version_outputs (
    state_version_output_id,
    name,
    sensitive,
    type,
    value,
    state_version_id
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Name'),
    pggen.arg('Sensitive'),
    pggen.arg('Type'),
    pggen.arg('Value'),
    pggen.arg('StateVersionID')
);
