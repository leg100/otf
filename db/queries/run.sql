-- name: InsertRun :one
INSERT INTO runs (
    run_id,
    plan_id,
    apply_id,
    created_at,
    updated_at,
    is_destroy,
    position_in_queue,
    refresh,
    refresh_only,
    status,
    plan_status,
    apply_status,
    replace_addrs,
    target_addrs,
    configuration_version_id,
    workspace_id
) VALUES (
    pggen.arg('ID'),
    pggen.arg('PlanID'),
    pggen.arg('ApplyID'),
    current_timestamp,
    current_timestamp,
    pggen.arg('IsDestroy'),
    pggen.arg('PositionInQueue'),
    pggen.arg('Refresh'),
    pggen.arg('RefreshOnly'),
    pggen.arg('Status'),
    pggen.arg('PlanStatus'),
    pggen.arg('ApplyStatus'),
    pggen.arg('ReplaceAddrs'),
    pggen.arg('TargetAddrs'),
    pggen.arg('ConfigurationVersionID'),
    pggen.arg('WorkspaceID')
)
RETURNING created_at, updated_at;

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

-- name: FindRuns :many
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.updated_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_changes,
    runs.applied_changes,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(pst.*) AS plan_status_timestamps
        FROM plan_status_timestamps pst
        WHERE pst.run_id = runs.run_id
        GROUP BY run_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(ast.*) AS apply_status_timestamps
        FROM apply_status_timestamps ast
        WHERE ast.run_id = runs.run_id
        GROUP BY run_id
    ) AS apply_status_timestamps
FROM runs
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
JOIN organizations USING(organization_id)
WHERE runs.workspace_id = pggen.arg('workspace_id')
AND runs.status LIKE ANY(pggen.arg('statuses'))
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: CountRuns :one
SELECT count(*)
FROM runs
WHERE workspace_id = pggen.arg('workspace_id')
AND status LIKE ANY(pggen.arg('statuses'))
;

-- name: FindRunByID :one
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.updated_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_changes,
    runs.applied_changes,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(pst.*) AS plan_status_timestamps
        FROM plan_status_timestamps pst
        WHERE pst.run_id = runs.run_id
        GROUP BY run_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(ast.*) AS apply_status_timestamps
        FROM apply_status_timestamps ast
        WHERE ast.run_id = runs.run_id
        GROUP BY run_id
    ) AS apply_status_timestamps
FROM runs
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.run_id = pggen.arg('run_id')
;

-- name: FindRunByPlanID :one
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.updated_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_changes,
    runs.applied_changes,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(pst.*) AS plan_status_timestamps
        FROM plan_status_timestamps pst
        WHERE pst.run_id = runs.run_id
        GROUP BY run_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(ast.*) AS apply_status_timestamps
        FROM apply_status_timestamps ast
        WHERE ast.run_id = runs.run_id
        GROUP BY run_id
    ) AS apply_status_timestamps
FROM runs
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.plan_id = pggen.arg('plan_id')
;

-- name: FindRunByApplyID :one
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.updated_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_changes,
    runs.applied_changes,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(pst.*) AS plan_status_timestamps
        FROM plan_status_timestamps pst
        WHERE pst.run_id = runs.run_id
        GROUP BY run_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(ast.*) AS apply_status_timestamps
        FROM apply_status_timestamps ast
        WHERE ast.run_id = runs.run_id
        GROUP BY run_id
    ) AS apply_status_timestamps
FROM runs
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.apply_id = pggen.arg('apply_id')
;

-- name: FindRunByIDForUpdate :one
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.updated_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_changes,
    runs.applied_changes,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(pst.*) AS plan_status_timestamps
        FROM plan_status_timestamps pst
        WHERE pst.run_id = runs.run_id
        GROUP BY run_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(ast.*) AS apply_status_timestamps
        FROM apply_status_timestamps ast
        WHERE ast.run_id = runs.run_id
        GROUP BY run_id
    ) AS apply_status_timestamps
FROM runs
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.run_id = pggen.arg('run_id')
FOR UPDATE
;

-- name: FindRunByPlanIDForUpdate :one
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.updated_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_changes,
    runs.applied_changes,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(pst.*) AS plan_status_timestamps
        FROM plan_status_timestamps pst
        WHERE pst.run_id = runs.run_id
        GROUP BY run_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(ast.*) AS apply_status_timestamps
        FROM apply_status_timestamps ast
        WHERE ast.run_id = runs.run_id
        GROUP BY run_id
    ) AS apply_status_timestamps
FROM runs
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.plan_id = pggen.arg('plan_id')
FOR UPDATE
;

-- name: FindRunByApplyIDForUpdate :one
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.updated_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_changes,
    runs.applied_changes,
    (configuration_versions.*)::"configuration_versions" AS configuration_version,
    (workspaces.*)::"workspaces" AS workspace,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(pst.*) AS plan_status_timestamps
        FROM plan_status_timestamps pst
        WHERE pst.run_id = runs.run_id
        GROUP BY run_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(ast.*) AS apply_status_timestamps
        FROM apply_status_timestamps ast
        WHERE ast.run_id = runs.run_id
        GROUP BY run_id
    ) AS apply_status_timestamps
FROM runs
JOIN configuration_versions USING(workspace_id)
JOIN workspaces USING(workspace_id)
WHERE runs.apply_id = pggen.arg('apply_id')
FOR UPDATE
;

-- name: UpdateRunStatus :one
UPDATE runs
SET
    status = pggen.arg('status'),
    updated_at = current_timestamp
WHERE run_id = pggen.arg('id')
RETURNING updated_at
;

-- name: UpdateRunPlannedChangesByRunID :exec
UPDATE runs
SET planned_changes = ROW(pggen.arg('additions'), pggen.arg('changes'), pggen.arg('destructions'))
WHERE run_id = pggen.arg('id')
;

-- name: UpdateRunAppliedChangesByApplyID :exec
UPDATE runs
SET applied_changes = ROW(pggen.arg('additions'), pggen.arg('changes'), pggen.arg('destructions'))
WHERE apply_id = pggen.arg('id')
;

-- name: DeleteRunByID :exec
DELETE
FROM runs
WHERE run_id = pggen.arg('run_id');
