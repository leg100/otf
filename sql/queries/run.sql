-- name: InsertRun :exec
INSERT INTO runs (
    run_id,
    plan_id,
    apply_id,
    created_at,
    is_destroy,
    position_in_queue,
    refresh,
    refresh_only,
    status,
    plan_status,
    apply_status,
    replace_addrs,
    target_addrs,
    planned_additions,
    planned_changes,
    planned_destructions,
    applied_additions,
    applied_changes,
    applied_destructions,
    configuration_version_id,
    workspace_id
) VALUES (
    pggen.arg('ID'),
    pggen.arg('PlanID'),
    pggen.arg('ApplyID'),
    pggen.arg('CreatedAt'),
    pggen.arg('IsDestroy'),
    pggen.arg('PositionInQueue'),
    pggen.arg('Refresh'),
    pggen.arg('RefreshOnly'),
    pggen.arg('Status'),
    pggen.arg('PlanStatus'),
    pggen.arg('ApplyStatus'),
    pggen.arg('ReplaceAddrs'),
    pggen.arg('TargetAddrs'),
    pggen.arg('PlannedAdditions'),
    pggen.arg('PlannedChanges'),
    pggen.arg('PlannedDestructions'),
    pggen.arg('AppliedAdditions'),
    pggen.arg('AppliedChanges'),
    pggen.arg('AppliedDestructions'),
    pggen.arg('ConfigurationVersionID'),
    pggen.arg('WorkspaceID')
);

-- name: InsertRunStatusTimestamp :exec
INSERT INTO run_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    pggen.arg('Timestamp')
);

-- name: FindRuns :many
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_additions,
    runs.planned_changes,
    runs.planned_destructions,
    runs.applied_additions,
    runs.applied_changes,
    runs.applied_destructions,
    runs.configuration_version_id,
    runs.workspace_id,
    configuration_versions.speculative,
    workspaces.auto_apply,
    CASE WHEN pggen.arg('include_configuration_version') THEN (configuration_versions.*)::"configuration_versions" END AS configuration_version,
    CASE WHEN pggen.arg('include_workspace') THEN (workspaces.*)::"workspaces" END AS workspace,
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
JOIN configuration_versions USING(configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations USING(organization_id)
WHERE
    organizations.name      LIKE ANY(pggen.arg('organization_names'))
AND workspaces.workspace_id LIKE ANY(pggen.arg('workspace_ids'))
AND workspaces.name         LIKE ANY(pggen.arg('workspace_names'))
AND runs.status             LIKE ANY(pggen.arg('statuses'))
ORDER BY runs.created_at ASC
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: CountRuns :one
SELECT count(*)
FROM runs
JOIN workspaces USING(workspace_id)
JOIN organizations USING(organization_id)
WHERE
    organizations.name      LIKE ANY(pggen.arg('organization_names'))
AND workspaces.workspace_id LIKE ANY(pggen.arg('workspace_ids'))
AND workspaces.name         LIKE ANY(pggen.arg('workspace_names'))
AND runs.status             LIKE ANY(pggen.arg('statuses'))
;

-- name: FindRunByID :one
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_additions,
    runs.planned_changes,
    runs.planned_destructions,
    runs.applied_additions,
    runs.applied_changes,
    runs.applied_destructions,
    runs.configuration_version_id,
    runs.workspace_id,
    configuration_versions.speculative,
    workspaces.auto_apply,
    CASE WHEN pggen.arg('include_configuration_version') THEN (configuration_versions.*)::"configuration_versions" END AS configuration_version,
    CASE WHEN pggen.arg('include_workspace') THEN (workspaces.*)::"workspaces" END AS workspace,
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
JOIN configuration_versions USING(configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
WHERE runs.run_id = pggen.arg('run_id')
;

-- name: FindRunIDByPlanID :one
SELECT run_id
FROM runs
WHERE plan_id = pggen.arg('plan_id')
;

-- name: FindRunIDByApplyID :one
SELECT run_id
FROM runs
WHERE apply_id = pggen.arg('apply_id')
;

-- name: FindRunByIDForUpdate :one
SELECT
    runs.run_id,
    runs.plan_id,
    runs.apply_id,
    runs.created_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    runs.plan_status,
    runs.apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.planned_additions,
    runs.planned_changes,
    runs.planned_destructions,
    runs.applied_additions,
    runs.applied_changes,
    runs.applied_destructions,
    runs.configuration_version_id,
    runs.workspace_id,
    configuration_versions.speculative,
    workspaces.auto_apply,
    CASE WHEN pggen.arg('include_configuration_version') THEN (configuration_versions.*)::"configuration_versions" END AS configuration_version,
    CASE WHEN pggen.arg('include_workspace') THEN (workspaces.*)::"workspaces" END AS workspace,
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
JOIN configuration_versions USING(configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
WHERE runs.run_id = pggen.arg('run_id')
FOR UPDATE
;

-- name: UpdateRunStatus :one
UPDATE runs
SET
    status = pggen.arg('status')
WHERE run_id = pggen.arg('id')
RETURNING run_id
;

-- name: UpdateRunPlannedChangesByPlanID :one
UPDATE runs
SET
    planned_additions = pggen.arg('additions'),
    planned_changes = pggen.arg('changes'),
    planned_destructions = pggen.arg('destructions')
WHERE plan_id = pggen.arg('plan_id')
RETURNING plan_id
;

-- name: UpdateRunAppliedChangesByApplyID :one
UPDATE runs
SET
    applied_additions = pggen.arg('additions'),
    applied_changes = pggen.arg('changes'),
    applied_destructions = pggen.arg('destructions')
WHERE apply_id = pggen.arg('apply_id')
RETURNING plan_id
;

-- name: DeleteRunByID :exec
DELETE
FROM runs
WHERE run_id = pggen.arg('run_id');
