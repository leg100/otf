-- name: InsertApplyStatusTimestamp :one
INSERT INTO apply_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    current_timestamp
)
RETURNING *;

-- name: UpdateApplyStatus :one
UPDATE runs
SET
    apply_status = pggen.arg('status'),
    updated_at = current_timestamp
WHERE apply_id = pggen.arg('id')
RETURNING updated_at
;
