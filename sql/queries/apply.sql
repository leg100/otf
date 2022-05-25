-- name: InsertApplyStatusTimestamp :exec
INSERT INTO apply_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    pggen.arg('Timestamp')
);

-- name: UpdateApplyStatus :one
UPDATE runs
SET
    apply_status = pggen.arg('status')
WHERE apply_id = pggen.arg('id')
RETURNING apply_id
;
