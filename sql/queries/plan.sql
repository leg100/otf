-- name: InsertPlanStatusTimestamp :exec
INSERT INTO plan_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    pggen.arg('Timestamp')
);

-- name: UpdatePlanStatus :one
UPDATE runs
SET
    plan_status = pggen.arg('status')
WHERE plan_id = pggen.arg('id')
RETURNING plan_id
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

-- name: UpdateRunPlanBinByPlanID :exec
UPDATE runs
SET plan_bin = pggen.arg('plan_bin')
WHERE plan_id = pggen.arg('plan_id')
;

-- name: UpdateRunPlanJSONByPlanID :exec
UPDATE runs
SET plan_json = pggen.arg('plan_json')
WHERE plan_id = pggen.arg('plan_id')
;
