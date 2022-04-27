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
