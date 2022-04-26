-- name: InsertRun :one
INSERT INTO runs (
    run_id,
    created_at,
    updated_at,
    is_destroy,
    position_in_queue,
    refresh,
    refresh_only,
    status,
    status_timestamps,
    replace_addrs,
    target_addrs,
    workspace_id
) VALUES (
    pggen.arg('ID'),
    NOW(),
    NOW(),
    pggen.arg('IsDestroy'),
    pggen.arg('PositionInQueue'),
    pggen.arg('Refresh'),
    pggen.arg('RefreshOnly'),
    pggen.arg('Status'),
    pggen.arg('StatusTimestamps'),
    pggen.arg('ReplaceAddrs'),
    pggen.arg('TargetAddrs'),
    pggen.arg('WorkspaceID')
)
RETURNING *;

-- name: FindRunsByWorkspaceID :many
SELECT runs.*,
    (plans.*)::"plans" AS plans,
    (applies.*)::"applies" AS applies,
    (configuration_versions.*)::"configuration_versions" AS configuration_versions,
    (workspaces.*)::"workspaces" AS workspaces,
    count(*) OVER() AS full_count
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE workspaces.workspace_id = pggen.arg('workspace_id')
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: FindRunsByWorkspaceName :many
SELECT runs.*,
    (plans.*)::"plans" AS plans,
    (applies.*)::"applies" AS applies,
    (configuration_versions.*)::"configuration_versions" AS configuration_versions,
    (workspaces.*)::"workspaces" AS workspaces,
    count(*) OVER() AS full_count
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
JOIN organizations USING(organization_id)
WHERE workspaces.name = pggen.arg('workspace_name')
AND organizations.name = pggen.arg('organization_name')
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: FindRunByID :one
SELECT runs.*,
    (plans.*)::"plans" AS plans,
    (applies.*)::"applies" AS applies,
    (configuration_versions.*)::"configuration_versions" AS configuration_versions,
    (workspaces.*)::"workspaces" AS workspaces
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.run_id = pggen.arg('run_id')
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: FindRunByPlanID :one
SELECT runs.*,
    (plans.*)::"plans" AS plans,
    (applies.*)::"applies" AS applies,
    (configuration_versions.*)::"configuration_versions" AS configuration_versions,
    (workspaces.*)::"workspaces" AS workspaces
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE plans.plan_id = pggen.arg('plan_id')
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: FindRunByApplyID :one
SELECT runs.*,
    (plans.*)::"plans" AS plans,
    (applies.*)::"applies" AS applies,
    (configuration_versions.*)::"configuration_versions" AS configuration_versions,
    (workspaces.*)::"workspaces" AS workspaces
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE applies.apply_id = pggen.arg('apply_id')
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: GetPlanFileByRunID :one
SELECT plans.plan_file
FROM runs
JOIN plans USING(run_id)
WHERE runs.run_id = pggen.arg('run_id')
;

-- name: GetPlanJSONByRunID :one
SELECT plans.plan_json
FROM runs
JOIN plans USING(run_id)
WHERE runs.run_id = pggen.arg('run_id')
;

-- name: PutPlanFileByRunID :exec
UPDATE plans
SET plan_file = pggen.arg('plan_file')
WHERE run_id = pggen.arg('run_id')
;

-- name: PutPlanJSONByRunID :exec
UPDATE plans
SET plan_json = pggen.arg('plan_json')
WHERE run_id = pggen.arg('run_id')
;

-- name: UpdateRunStatus :one
UPDATE runs
SET
    status = pggen.arg('status'),
    status_timestamps = pggen.arg('status_timestamps'),
    updated_at = NOW()
WHERE run_id = pggen.arg('id')
RETURNING *;

-- name: DeleteRunByID :exec
DELETE
FROM runs
WHERE run_id = pggen.arg('run_id');
