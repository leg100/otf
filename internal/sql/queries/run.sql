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
    auto_apply,
    configuration_version_id,
    workspace_id
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('is_destroy'),
    pggen.arg('position_in_queue'),
    pggen.arg('refresh'),
    pggen.arg('refresh_only'),
    pggen.arg('status'),
    pggen.arg('replace_addrs'),
    pggen.arg('target_addrs'),
    pggen.arg('auto_apply'),
    pggen.arg('configuration_version_id'),
    pggen.arg('workspace_id')
);

-- name: InsertRunStatusTimestamp :exec
INSERT INTO run_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    pggen.arg('id'),
    pggen.arg('status'),
    pggen.arg('timestamp')
);

-- name: FindRuns :many
SELECT
    runs.run_id,
    runs.created_at,
    runs.force_cancel_available_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    plans.status      AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.report AS planned_changes,
    applies.report AS applied_changes,
    runs.configuration_version_id,
    runs.workspace_id,
    configuration_versions.speculative,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    (ia.*)::"ingress_attributes" AS ingress_attributes,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(st.*) AS phase_status_timestamps
        FROM phase_status_timestamps st
        WHERE st.run_id = plans.run_id
        AND   st.phase = 'plan'
        GROUP BY run_id, phase
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(st.*) AS phase_status_timestamps
        FROM phase_status_timestamps st
        WHERE st.run_id = applies.run_id
        AND   st.phase = 'apply'
        GROUP BY run_id, phase
    ) AS apply_status_timestamps
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
WHERE
    workspaces.organization_name             LIKE ANY(pggen.arg('organization_names'))
AND workspaces.workspace_id                  LIKE ANY(pggen.arg('workspace_ids'))
AND workspaces.name                          LIKE ANY(pggen.arg('workspace_names'))
AND runs.status                              LIKE ANY(pggen.arg('statuses'))
AND configuration_versions.speculative::text LIKE ANY(pggen.arg('speculative'))
ORDER BY runs.created_at DESC
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: CountRuns :one
SELECT count(*)
FROM runs
JOIN workspaces             USING(workspace_id)
JOIN configuration_versions USING(configuration_version_id)
WHERE
    workspaces.organization_name             LIKE ANY(pggen.arg('organization_names'))
AND workspaces.workspace_id                  LIKE ANY(pggen.arg('workspace_ids'))
AND workspaces.name                          LIKE ANY(pggen.arg('workspace_names'))
AND runs.status                              LIKE ANY(pggen.arg('statuses'))
AND configuration_versions.speculative::text LIKE ANY(pggen.arg('speculative'))
;

-- name: FindRunByID :one
SELECT
    runs.run_id,
    runs.created_at,
    runs.force_cancel_available_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    plans.status      AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.report AS planned_changes,
    applies.report AS applied_changes,
    runs.configuration_version_id,
    runs.workspace_id,
    configuration_versions.speculative,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    (ia.*)::"ingress_attributes" AS ingress_attributes,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(st.*) AS phase_status_timestamps
        FROM phase_status_timestamps st
        WHERE st.run_id = plans.run_id
        AND   st.phase = 'plan'
        GROUP BY run_id, phase
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(st.*) AS phase_status_timestamps
        FROM phase_status_timestamps st
        WHERE st.run_id = applies.run_id
        AND   st.phase = 'apply'
        GROUP BY run_id, phase
    ) AS apply_status_timestamps
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
WHERE runs.run_id = pggen.arg('run_id')
;

-- name: FindRunByIDForUpdate :one
SELECT
    runs.run_id,
    runs.created_at,
    runs.force_cancel_available_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.status,
    plans.status        AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.report AS planned_changes,
    applies.report AS applied_changes,
    runs.configuration_version_id,
    runs.workspace_id,
    configuration_versions.speculative,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    (ia.*)::"ingress_attributes" AS ingress_attributes,
    (
        SELECT array_agg(rst.*) AS run_status_timestamps
        FROM run_status_timestamps rst
        WHERE rst.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_status_timestamps,
    (
        SELECT array_agg(st.*) AS phase_status_timestamps
        FROM phase_status_timestamps st
        WHERE st.run_id = plans.run_id
        AND   st.phase = 'plan'
        GROUP BY run_id, phase
    ) AS plan_status_timestamps,
    (
        SELECT array_agg(st.*) AS phase_status_timestamps
        FROM phase_status_timestamps st
        WHERE st.run_id = applies.run_id
        AND   st.phase = 'apply'
        GROUP BY run_id, phase
    ) AS apply_status_timestamps
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
WHERE runs.run_id = pggen.arg('run_id')
FOR UPDATE of runs, plans, applies
;

-- name: PutLockFile :one
UPDATE runs
SET lock_file = pggen.arg('lock_file')
WHERE run_id = pggen.arg('run_id')
RETURNING run_id
;

-- name: GetLockFileByID :one
SELECT lock_file
FROM runs
WHERE run_id = pggen.arg('run_id')
;

-- name: UpdateRunStatus :one
UPDATE runs
SET
    status = pggen.arg('status')
WHERE run_id = pggen.arg('id')
RETURNING run_id
;

-- name: UpdateRunForceCancelAvailableAt :one
UPDATE runs
SET
    force_cancel_available_at = pggen.arg('force_cancel_available_at')
WHERE run_id = pggen.arg('id')
RETURNING run_id
;

-- name: DeleteRunByID :one
DELETE
FROM runs
WHERE run_id = pggen.arg('run_id')
RETURNING run_id
;
