-- name: InsertPlan :exec
INSERT INTO plans (
    plan_id,
    run_id,
    status
) VALUES (
    pggen.arg('plan_id'),
    pggen.arg('run_id'),
    pggen.arg('status')
);

-- name: InsertPlanStatusTimestamp :exec
INSERT INTO plan_status_timestamps (
    plan_id,
    status,
    timestamp
) VALUES (
    pggen.arg('plan_id'),
    pggen.arg('status'),
    pggen.arg('timestamp')
);

-- name: UpdatePlanStatusByID :one
UPDATE plans
SET status = pggen.arg('status')
WHERE plan_id = pggen.arg('plan_id')
RETURNING plan_id
;

-- name: UpdatePlannedChangesByID :one
UPDATE plans
SET report = (
    pggen.arg('additions'),
    pggen.arg('changes'),
    pggen.arg('destructions')
)
WHERE plan_id = pggen.arg('plan_id')
RETURNING plan_id
;

-- name: FindRunIDByPlanID :one
SELECT run_id
FROM plans
WHERE plan_id = pggen.arg('plan_id')
;

-- name: GetPlanBinByID :one
SELECT plan_bin
FROM plans
WHERE plan_id = pggen.arg('plan_id')
;

-- name: GetPlanJSONByID :one
SELECT plan_json
FROM plans
WHERE plan_id = pggen.arg('plan_id')
;

-- name: UpdatePlanBinByID :one
UPDATE plans
SET plan_bin = pggen.arg('plan_bin')
WHERE plan_id = pggen.arg('plan_id')
RETURNING plan_id
;

-- name: UpdatePlanJSONByID :one
UPDATE plans
SET plan_json = pggen.arg('plan_json')
WHERE plan_id = pggen.arg('plan_id')
RETURNING plan_id
;

-- name: InsertPlanLogChunk :exec
INSERT INTO plan_logs (
    plan_id,
    chunk
) VALUES (
    pggen.arg('plan_id'),
    pggen.arg('chunk')
)
;

-- name: FindPlanLogChunks :one
SELECT
    substring(string_agg(chunk, '') FROM pggen.arg('offset') FOR pggen.arg('limit'))
FROM (
    SELECT plan_id, chunk
    FROM plan_logs
    WHERE plan_id = pggen.arg('plan_id')
    ORDER BY chunk_id
) c
GROUP BY plan_id
;
