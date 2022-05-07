-- name: InsertApply :one
INSERT INTO applies (
    apply_id,
    created_at,
    updated_at,
    status,
    run_id
) VALUES (
    pggen.arg('ID'),
    NOW(),
    NOW(),
    pggen.arg('Status'),
    pggen.arg('RunID')
)
RETURNING *;

-- name: InsertApplyStatusTimestamp :one
INSERT INTO apply_status_timestamps (
    apply_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    NOW()
)
RETURNING *;

-- name: UpdateApplyStatus :one
UPDATE applies
SET
    status = pggen.arg('status'),
    updated_at = NOW()
WHERE apply_id = pggen.arg('id')
RETURNING *;

-- name: UpdateApplyResources :exec
UPDATE applies
SET
    resource_additions = pggen.arg('resource_additions'),
    resource_changes = pggen.arg('resource_changes'),
    resource_destructions = pggen.arg('resource_destructions')
WHERE run_id = pggen.arg('run_id')
;
