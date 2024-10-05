-- name: InsertPlan :exec
INSERT INTO plans (
    run_id,
    status
) VALUES (
    sqlc.arg('run_id'),
    sqlc.arg('status')
);

-- name: UpdatePlanStatusByID :one
UPDATE plans
SET status = sqlc.arg('status')
WHERE run_id = sqlc.arg('run_id')
RETURNING run_id
;

-- name: UpdatePlannedChangesByID :one
UPDATE plans
SET resource_report = (
        sqlc.arg('resource_additions'),
        sqlc.arg('resource_changes'),
        sqlc.arg('resource_destructions')
    ),
    output_report = (
        sqlc.arg('output_additions'),
        sqlc.arg('output_changes'),
        sqlc.arg('output_destructions')
    )
WHERE run_id = sqlc.arg('run_id')
RETURNING run_id
;

-- name: GetPlanBinByID :one
SELECT plan_bin
FROM plans
WHERE run_id = sqlc.arg('run_id')
;

-- name: GetPlanJSONByID :one
SELECT plan_json
FROM plans
WHERE run_id = sqlc.arg('run_id')
;

-- name: UpdatePlanBinByID :one
UPDATE plans
SET plan_bin = sqlc.arg('plan_bin')
WHERE run_id = sqlc.arg('run_id')
RETURNING run_id
;

-- name: UpdatePlanJSONByID :one
UPDATE plans
SET plan_json = sqlc.arg('plan_json')
WHERE run_id = sqlc.arg('run_id')
RETURNING run_id
;
