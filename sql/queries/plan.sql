-- name: InsertPlan :exec
INSERT INTO plans (
    plan_id,
    job_id,
    additions,
    changes,
    destructions
) VALUES (
    pggen.arg('plan_id'),
    pggen.arg('job_id'),
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

-- name: FindRunIDByPlanID :one
SELECT jobs.run_id
FROM plans
JOIN jobs USING(job_id)
WHERE plans.plan_id = pggen.arg('plan_id')
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
