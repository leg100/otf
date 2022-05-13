-- name: InsertPlanStatusTimestamp :one
INSERT INTO plan_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    current_timestamp
)
RETURNING *;

-- name: UpdatePlanStatus :one
UPDATE runs
SET
    status = pggen.arg('status'),
    updated_at = current_timestamp
WHERE plan_id = pggen.arg('id')
RETURNING *;

-- name: GetPlanBinByRunID :one
SELECT plan_bin
FROM runs
WHERE run_id = pggen.arg('run_id')
;

-- name: GetPlanJSONByRunID :one
SELECT plan_json
FROM runs
WHERE run_id = pggen.arg('run_id')
;

-- name: PutPlanBinByRunID :exec
UPDATE runs
SET plan_bin = pggen.arg('plan_bin')
WHERE run_id = pggen.arg('run_id')
;

-- name: PutPlanJSONByRunID :exec
UPDATE runs
SET plan_json = pggen.arg('plan_json')
WHERE run_id = pggen.arg('run_id')
;

-- name: UpdatePlanResources :exec
UPDATE runs
SET
    planned_resource_additions = pggen.arg('planned_resource_additions'),
    planned_resource_changes = pggen.arg('planned_resource_changes'),
    planned_resource_destructions = pggen.arg('planned_resource_destructions')
WHERE plan_id = pggen.arg('plan_id')
;
