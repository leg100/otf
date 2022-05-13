-- name: InsertPlan :one
INSERT INTO plans (
    plan_id,
    created_at,
    updated_at,
    status,
    run_id
) VALUES (
    pggen.arg('ID'),
    current_timestamp,
    current_timestamp,
    pggen.arg('Status'),
    pggen.arg('RunID')
)
RETURNING *;

-- name: InsertPlanStatusTimestamp :one
INSERT INTO plan_status_timestamps (
    plan_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    current_timestamp
)
RETURNING *;

-- name: UpdatePlanStatus :one
UPDATE plans
SET
    status = pggen.arg('status'),
    updated_at = current_timestamp
WHERE plan_id = pggen.arg('id')
RETURNING *;

-- name: GetPlanBinByRunID :one
SELECT plan_bin
FROM plans
WHERE run_id = pggen.arg('run_id')
;

-- name: GetPlanJSONByRunID :one
SELECT plan_json
FROM plans
WHERE run_id = pggen.arg('run_id')
;

-- name: PutPlanBinByRunID :exec
UPDATE plans
SET plan_bin = pggen.arg('plan_bin')
WHERE run_id = pggen.arg('run_id')
;

-- name: PutPlanJSONByRunID :exec
UPDATE plans
SET plan_json = pggen.arg('plan_json')
WHERE run_id = pggen.arg('run_id')
;

-- name: UpdatePlanResources :exec
UPDATE plans
SET
    resource_additions = pggen.arg('resource_additions'),
    resource_changes = pggen.arg('resource_changes'),
    resource_destructions = pggen.arg('resource_destructions')
WHERE run_id = pggen.arg('run_id')
;
