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
    status = pggen.arg('status'),
    updated_at = current_timestamp
WHERE apply_id = pggen.arg('id')
RETURNING updated_at
;

-- name: InsertApplyResourceReport :exec
INSERT INTO apply_resource_reports (
    apply_id,
    resource_additions,
    resource_changes,
    resource_destructions
) VALUES (
    pggen.arg('apply_id'),
    pggen.arg('additions'),
    pggen.arg('changes'),
    pggen.arg('destructions')
)
;
