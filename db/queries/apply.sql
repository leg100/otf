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
RETURNING *;

-- name: UpdateApplyResources :exec
UPDATE runs
SET
    applied_resource_additions = pggen.arg('applied_resource_additions'),
    applied_resource_changes = pggen.arg('applied_resource_changes'),
    applied_resource_destructions = pggen.arg('applied_resource_destructions')
WHERE apply_id = pggen.arg('apply_id')
;
