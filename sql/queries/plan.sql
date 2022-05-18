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
    plan_status = pggen.arg('status'),
    updated_at = current_timestamp
WHERE plan_id = pggen.arg('id')
RETURNING updated_at
;

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
