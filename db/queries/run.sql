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
    replace_addrs,
    target_addrs,
    configuration_version_id,
    workspace_id
) VALUES (
    pggen.arg('ID'),
    current_timestamp,
    current_timestamp,
    pggen.arg('IsDestroy'),
    pggen.arg('PositionInQueue'),
    pggen.arg('Refresh'),
    pggen.arg('RefreshOnly'),
    pggen.arg('Status'),
    pggen.arg('ReplaceAddrs'),
    pggen.arg('TargetAddrs'),
    pggen.arg('ConfigurationVersionID'),
    pggen.arg('WorkspaceID')
)
RETURNING *;

-- name: InsertRunStatusTimestamp :one
INSERT INTO run_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    current_timestamp
)
RETURNING *;

-- name: FindRunsByWorkspaceID :many
SELECT
    runs.*,
    (plans.*)::"plans" AS plan,
    (applies.*)::"applies" AS apply,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
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
SELECT
    runs.*,
    (plans.*)::"plans" AS plan,
    (applies.*)::"applies" AS apply,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
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

-- name: FindRunsByStatuses :many
SELECT runs.*,
    (plans.*)::"plans" AS plan,
    (applies.*)::"applies" AS apply,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    count(*) OVER() AS full_count
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.status = ANY(pggen.arg('statuses'))
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: FindRunByID :one
SELECT runs.*,
    (plans.*)::"plans" AS plan,
    (applies.*)::"applies" AS apply,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.run_id = pggen.arg('run_id')
;

-- name: FindRunByPlanID :one
SELECT runs.*,
    (plans.*)::"plans" AS plan,
    (applies.*)::"applies" AS apply,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE plans.plan_id = pggen.arg('plan_id')
;

-- name: FindRunByApplyID :one
SELECT runs.*,
    (plans.*)::"plans" AS plan,
    (applies.*)::"applies" AS apply,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE applies.apply_id = pggen.arg('apply_id')
;

-- name: FindRunByIDForUpdate :one
SELECT runs.*,
    (plans.*)::"plans" AS plan,
    (applies.*)::"applies" AS apply,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.run_id = pggen.arg('run_id')
FOR UPDATE
;

-- name: UpdateRunStatus :one
UPDATE runs
SET
    status = pggen.arg('status'),
    updated_at = current_timestamp
WHERE run_id = pggen.arg('id')
RETURNING *;

-- name: DeleteRunByID :exec
DELETE
FROM runs
WHERE run_id = pggen.arg('run_id');
