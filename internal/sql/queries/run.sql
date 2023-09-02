-- name: InsertRun :exec
INSERT INTO runs (
    run_id,
    created_at,
    is_destroy,
    position_in_queue,
    refresh,
    refresh_only,
    source,
    status,
    replace_addrs,
    target_addrs,
    auto_apply,
    plan_only,
    configuration_version_id,
    workspace_id,
    created_by,
    terraform_version,
    allow_empty_apply
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('is_destroy'),
    pggen.arg('position_in_queue'),
    pggen.arg('refresh'),
    pggen.arg('refresh_only'),
    pggen.arg('source'),
    pggen.arg('status'),
    pggen.arg('replace_addrs'),
    pggen.arg('target_addrs'),
    pggen.arg('auto_apply'),
    pggen.arg('plan_only'),
    pggen.arg('configuration_version_id'),
    pggen.arg('workspace_id'),
    pggen.arg('created_by'),
    pggen.arg('terraform_version'),
    pggen.arg('allow_empty_apply')
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

-- name: InsertRunVariable :exec
INSERT INTO run_variables (
    run_id,
    key,
    value
) VALUES (
    pggen.arg('run_id'),
    pggen.arg('key'),
    pggen.arg('value')
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
    runs.source,
    runs.status,
    plans.status      AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.resource_report AS plan_resource_report,
    plans.output_report AS plan_output_report,
    applies.resource_report AS apply_resource_report,
    runs.configuration_version_id,
    runs.workspace_id,
    runs.plan_only,
    runs.created_by,
    runs.terraform_version,
    runs.allow_empty_apply,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    organizations.cost_estimation_enabled,
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
    ) AS apply_status_timestamps,
    (
        SELECT array_agg(v.*) AS run_variables
        FROM run_variables v
        WHERE v.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_variables
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
WHERE
    workspaces.organization_name LIKE ANY(pggen.arg('organization_names'))
AND workspaces.workspace_id      LIKE ANY(pggen.arg('workspace_ids'))
AND workspaces.name              LIKE ANY(pggen.arg('workspace_names'))
AND runs.source                  LIKE ANY(pggen.arg('sources'))
AND runs.status                  LIKE ANY(pggen.arg('statuses'))
AND runs.plan_only::text         LIKE ANY(pggen.arg('plan_only'))
AND ((pggen.arg('commit_sha')::text IS NULL) OR ia.commit_sha = pggen.arg('commit_sha'))
AND ((pggen.arg('vcs_username')::text IS NULL) OR ia.sender_username = pggen.arg('vcs_username'))
ORDER BY runs.created_at DESC
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: CountRuns :one
SELECT count(*)
FROM runs
JOIN workspaces USING(workspace_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
WHERE
    workspaces.organization_name LIKE ANY(pggen.arg('organization_names'))
AND workspaces.workspace_id      LIKE ANY(pggen.arg('workspace_ids'))
AND workspaces.name              LIKE ANY(pggen.arg('workspace_names'))
AND runs.source                  LIKE ANY(pggen.arg('sources'))
AND runs.status                  LIKE ANY(pggen.arg('statuses'))
AND runs.plan_only::text         LIKE ANY(pggen.arg('plan_only'))
AND ((pggen.arg('commit_sha')::text IS NULL) OR ia.commit_sha = pggen.arg('commit_sha'))
AND ((pggen.arg('vcs_username')::text IS NULL) OR ia.sender_username = pggen.arg('vcs_username'))
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
    runs.source,
    runs.status,
    plans.status      AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.resource_report AS plan_resource_report,
    plans.output_report AS plan_output_report,
    applies.resource_report AS apply_resource_report,
    runs.configuration_version_id,
    runs.workspace_id,
    runs.plan_only,
    runs.created_by,
    runs.terraform_version,
    runs.allow_empty_apply,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    organizations.cost_estimation_enabled,
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
    ) AS apply_status_timestamps,
    (
        SELECT array_agg(v.*) AS run_variables
        FROM run_variables v
        WHERE v.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_variables
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
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
    runs.source,
    runs.status,
    plans.status        AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.resource_report AS plan_resource_report,
    plans.output_report AS plan_output_report,
    applies.resource_report AS apply_resource_report,
    runs.configuration_version_id,
    runs.workspace_id,
    runs.plan_only,
    runs.created_by,
    runs.terraform_version,
    runs.allow_empty_apply,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    organizations.cost_estimation_enabled,
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
    ) AS apply_status_timestamps,
    (
        SELECT array_agg(v.*) AS run_variables
        FROM run_variables v
        WHERE v.run_id = runs.run_id
        GROUP BY run_id
    ) AS run_variables
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
WHERE runs.run_id = pggen.arg('run_id')
FOR UPDATE OF runs, plans, applies
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
