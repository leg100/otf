-- name: InsertPlan :exec
INSERT INTO plans (
    plan_id,
    job_id,
    run_id,
    status,
    additions,
    changes,
    destructions
) VALUES (
    pggen.arg('plan_id'),
    pggen.arg('job_id'),
    pggen.arg('run_id'),
    pggen.arg('status'),
    pggen.arg('additions'),
    pggen.arg('changes'),
    pggen.arg('destructions')
);

-- name: UpdatePlannedChangesByID :one
UPDATE plans
SET
    additions = pggen.arg('additions'),
    changes = pggen.arg('changes'),
    destructions = pggen.arg('destructions')
WHERE plan_id = pggen.arg('plan_id')
RETURNING plan_id
;

-- name: FindPlanByID :one
SELECT
    plan_id,
    status,
    additions,
    changes,
    destructions,
    (
        SELECT array_agg(st.*)
        FROM job_status_timestamps st
        WHERE st.job_id = p.job_id
        GROUP BY p.job_id
    ) AS status_timestamps
FROM plans p
WHERE plan_id = pggen.arg('plan_id')
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
