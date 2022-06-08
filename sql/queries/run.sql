-- name: InsertRun :exec
INSERT INTO runs (
    run_id,
    created_at,
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
    pggen.arg('CreatedAt'),
    pggen.arg('IsDestroy'),
    pggen.arg('PositionInQueue'),
    pggen.arg('Refresh'),
    pggen.arg('RefreshOnly'),
    pggen.arg('Status'),
    pggen.arg('ReplaceAddrs'),
    pggen.arg('TargetAddrs'),
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
    plans.plan_id,
    applies.apply_id,
    runs.created_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    plans.status        AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    plans.additions      AS planned_additions,
    plans.changes        AS planned_changes,
    plans.destructions   AS planned_destructions,
    applies.additions    AS applied_additions,
    applies.changes      AS applied_changes,
    applies.destructions AS applied_destructions,
    runs.configuration_version_id,
    runs.workspace_id,
    configuration_versions.speculative,
    workspaces.auto_apply,
    workspaces.name AS workspace_name,
    organizations.name AS organization_name,
    CASE WHEN pggen.arg('include_configuration_version') THEN (configuration_versions.*)::"configuration_versions" END AS configuration_version,
    CASE WHEN pggen.arg('include_workspace') THEN (workspaces.*)::"workspaces" END AS workspace,
    (
        SELECT array_agg(st.*) AS run_status_timestamps
        FROM run_status_timestamps st
        WHERE st.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(st.*) AS plan_status_timestamps
        FROM job_status_timestamps st
        WHERE st.job_id = plans.job_id
        GROUP BY job_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(st.*) AS apply_status_timestamps
        FROM job_status_timestamps st
        WHERE st.job_id = applies.job_id
        GROUP BY job_id
    ) AS apply_status_timestamps
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
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
    plans.plan_id,
    applies.apply_id,
    runs.created_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    plans.status        AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    plans.additions      AS planned_additions,
    plans.changes        AS planned_changes,
    plans.destructions   AS planned_destructions,
    applies.additions    AS applied_additions,
    applies.changes      AS applied_changes,
    applies.destructions AS applied_destructions,
    runs.configuration_version_id,
    runs.workspace_id,
    configuration_versions.speculative,
    workspaces.auto_apply,
    workspaces.name AS workspace_name,
    organizations.name AS organization_name,
    CASE WHEN pggen.arg('include_configuration_version') THEN (configuration_versions.*)::"configuration_versions" END AS configuration_version,
    CASE WHEN pggen.arg('include_workspace') THEN (workspaces.*)::"workspaces" END AS workspace,
    (
        SELECT array_agg(st.*) AS run_status_timestamps
        FROM run_status_timestamps st
        WHERE st.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(st.*) AS plan_status_timestamps
        FROM job_status_timestamps st
        WHERE st.job_id = plans.job_id
        GROUP BY job_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(st.*) AS apply_status_timestamps
        FROM job_status_timestamps st
        WHERE st.job_id = applies.job_id
        GROUP BY job_id
    ) AS apply_status_timestamps
FROM runs
JOIN plans USING(run_id)
JOIN applies USING(run_id)
JOIN configuration_versions USING(configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations USING(organization_id)
WHERE runs.run_id = pggen.arg('run_id')
;

-- name: FindRunByIDForUpdate :one
SELECT
    runs.run_id,
    plans.plan_id,
    applies.apply_id,
    plan_jobs.job_id AS plan_job_id,
    apply_jobs.job_id AS apply_job_id,
    runs.created_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    plans.status        AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    plans.additions      AS planned_additions,
    plans.changes        AS planned_changes,
    plans.destructions   AS planned_destructions,
    applies.additions    AS applied_additions,
    applies.changes      AS applied_changes,
    applies.destructions AS applied_destructions,
    runs.configuration_version_id,
    runs.workspace_id,
    configuration_versions.speculative,
    workspaces.auto_apply,
    workspaces.name AS workspace_name,
    organizations.name AS organization_name,
    CASE WHEN pggen.arg('include_configuration_version') THEN (configuration_versions.*)::"configuration_versions" END AS configuration_version,
    CASE WHEN pggen.arg('include_workspace') THEN (workspaces.*)::"workspaces" END AS workspace,
    (
        SELECT array_agg(st.*) AS run_status_timestamps
        FROM run_status_timestamps st
        WHERE st.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(st.*) AS plan_status_timestamps
        FROM job_status_timestamps st
        WHERE st.job_id = plans.job_id
        GROUP BY job_id
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(st.*) AS apply_status_timestamps
        FROM job_status_timestamps st
        WHERE st.job_id = applies.job_id
        GROUP BY job_id
    ) AS apply_status_timestamps
FROM runs
JOIN (plans JOIN jobs plan_jobs USING (job_id)) ON plans.run_id = runs.run_id
JOIN (applies JOIN jobs apply_jobs USING (job_id)) ON applies.run_id = runs.run_id
JOIN configuration_versions USING(configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations USING(organization_id)
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

-- name: DeleteRunByID :one
DELETE
FROM runs
WHERE run_id = pggen.arg('run_id')
RETURNING run_id
;
