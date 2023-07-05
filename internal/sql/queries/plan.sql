-- name: InsertPlan :exec
INSERT INTO plans (
    run_id,
    status
) VALUES (
    pggen.arg('run_id'),
    pggen.arg('status')
);

-- name: UpdatePlanStatusByID :one
UPDATE plans
SET status = pggen.arg('status')
WHERE run_id = pggen.arg('run_id')
RETURNING run_id
;

-- name: UpdatePlannedChangesByID :one
UPDATE plans
SET resource_report = (
        pggen.arg('resource_additions'),
        pggen.arg('resource_changes'),
        pggen.arg('resource_destructions')
    ),
    output_report = (
        pggen.arg('output_additions'),
        pggen.arg('output_changes'),
        pggen.arg('output_destructions')
    )
WHERE run_id = pggen.arg('run_id')
RETURNING run_id
;

-- name: GetPlanBinByID :one
SELECT plan_bin
FROM plans
WHERE run_id = pggen.arg('run_id')
;

-- name: GetPlanJSONByID :one
SELECT plan_json
FROM plans
WHERE run_id = pggen.arg('run_id')
;

-- name: UpdatePlanBinByID :one
UPDATE plans
SET plan_bin = pggen.arg('plan_bin')
WHERE run_id = pggen.arg('run_id')
RETURNING run_id
;

-- name: UpdatePlanJSONByID :one
UPDATE plans
SET plan_json = pggen.arg('plan_json')
WHERE run_id = pggen.arg('run_id')
RETURNING run_id
;
