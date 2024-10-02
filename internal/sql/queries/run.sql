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
    sqlc.arg('id'),
    sqlc.arg('created_at'),
    sqlc.arg('is_destroy'),
    sqlc.arg('position_in_queue'),
    sqlc.arg('refresh'),
    sqlc.arg('refresh_only'),
    sqlc.arg('source'),
    sqlc.arg('status'),
    sqlc.arg('replace_addrs'),
    sqlc.arg('target_addrs'),
    sqlc.arg('auto_apply'),
    sqlc.arg('plan_only'),
    sqlc.arg('configuration_version_id'),
    sqlc.arg('workspace_id'),
    sqlc.arg('created_by'),
    sqlc.arg('terraform_version'),
    sqlc.arg('allow_empty_apply')
);

-- name: InsertRunStatusTimestamp :exec
INSERT INTO run_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    sqlc.arg('id'),
    sqlc.arg('status'),
    sqlc.arg('timestamp')
);

-- name: InsertRunVariable :exec
INSERT INTO run_variables (
    run_id,
    key,
    value
) VALUES (
    sqlc.arg('run_id'),
    sqlc.arg('key'),
    sqlc.arg('value')
);

-- name: FindRuns :many
SELECT
    runs.run_id,
    runs.created_at,
    runs.cancel_signaled_at,
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
    plans.resource_report::"report" AS plan_resource_report,
    plans.output_report::"report" AS plan_output_report,
    applies.resource_report::"report" AS apply_resource_report,
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
    rst.run_status_timestamps,
    pst.plan_status_timestamps,
    ast.apply_status_timestamps,
    rv.run_variables,
    ia::"ingress_attributes" AS ingress_attributes
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
JOIN (configuration_versions cv LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rv.*)::run_variables[] AS run_variables
    FROM run_variables rv
    GROUP BY run_id
) AS rv ON rv.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rst.*)::run_status_timestamps[] AS run_status_timestamps
    FROM run_status_timestamps rst
    GROUP BY run_id
) AS rst ON rst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(pst.*)::phase_status_timestamps[] AS plan_status_timestamps
    FROM phase_status_timestamps pst
    WHERE pst.phase = 'plan'
    GROUP BY run_id
) AS pst ON pst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(ast.*)::phase_status_timestamps[] AS apply_status_timestamps
    FROM phase_status_timestamps ast
    WHERE ast.phase = 'apply'
    GROUP BY run_id
) AS ast ON ast.run_id = runs.run_id
WHERE
    workspaces.organization_name LIKE ANY(sqlc.arg('organization_names')::text[])
AND workspaces.workspace_id      LIKE ANY(sqlc.arg('workspace_ids')::text[])
AND workspaces.name              LIKE ANY(sqlc.arg('workspace_names')::text[])
AND runs.source                  LIKE ANY(sqlc.arg('sources')::text[])
AND runs.status                  LIKE ANY(sqlc.arg('statuses')::text[])
AND runs.plan_only::text         LIKE ANY(sqlc.arg('plan_only')::text[])
AND ((sqlc.arg('commit_sha')::text IS NULL) OR ia.commit_sha = sqlc.arg('commit_sha'))
AND ((sqlc.arg('vcs_username')::text IS NULL) OR ia.sender_username = sqlc.arg('vcs_username'))
ORDER BY runs.created_at DESC
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int
;

-- name: CountRuns :one
SELECT count(*)
FROM runs
JOIN workspaces USING(workspace_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
WHERE
    workspaces.organization_name LIKE ANY(sqlc.arg('organization_names')::text[])
AND workspaces.workspace_id      LIKE ANY(sqlc.arg('workspace_ids')::text[])
AND workspaces.name              LIKE ANY(sqlc.arg('workspace_names')::text[])
AND runs.source                  LIKE ANY(sqlc.arg('sources')::text[])
AND runs.status                  LIKE ANY(sqlc.arg('statuses')::text[])
AND runs.plan_only::text         LIKE ANY(sqlc.arg('plan_only')::text[])
AND ((sqlc.arg('commit_sha')::text IS NULL) OR ia.commit_sha = sqlc.arg('commit_sha'))
AND ((sqlc.arg('vcs_username')::text IS NULL) OR ia.sender_username = sqlc.arg('vcs_username'))
;

-- name: FindRunByID :one
SELECT
    runs.run_id,
    runs.created_at,
    runs.cancel_signaled_at,
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
    plans.resource_report::"report" AS plan_resource_report,
    plans.output_report::"report" AS plan_output_report,
    applies.resource_report::"report" AS apply_resource_report,
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
    rst.run_status_timestamps,
    pst.plan_status_timestamps,
    ast.apply_status_timestamps,
    rv.run_variables,
    ia::"ingress_attributes" AS ingress_attributes
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
JOIN (configuration_versions cv LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rv.*)::run_variables[] AS run_variables
    FROM run_variables rv
    GROUP BY run_id
) AS rv ON rv.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rst.*)::run_status_timestamps[] AS run_status_timestamps
    FROM run_status_timestamps rst
    GROUP BY run_id
) AS rst ON rst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(pst.*)::phase_status_timestamps[] AS plan_status_timestamps
    FROM phase_status_timestamps pst
    WHERE pst.phase = 'plan'
    GROUP BY run_id
) AS pst ON pst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(ast.*)::phase_status_timestamps[] AS apply_status_timestamps
    FROM phase_status_timestamps ast
    WHERE ast.phase = 'apply'
    GROUP BY run_id
) AS ast ON ast.run_id = runs.run_id
WHERE runs.run_id = sqlc.arg('run_id')
GROUP BY runs.run_id
;

-- name: FindRunByIDForUpdate :one
SELECT
    runs.run_id,
    runs.created_at,
    runs.cancel_signaled_at,
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
    plans.resource_report::"report" AS plan_resource_report,
    plans.output_report::"report" AS plan_output_report,
    applies.resource_report::"report" AS apply_resource_report,
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
    rst.run_status_timestamps,
    pst.plan_status_timestamps,
    ast.apply_status_timestamps,
    rv.run_variables,
    ia::"ingress_attributes" AS ingress_attributes
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rv.*)::run_variables[] AS run_variables
    FROM run_variables rv
    GROUP BY run_id
) AS rv ON rv.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rst.*)::run_status_timestamps[] AS run_status_timestamps
    FROM run_status_timestamps rst
    GROUP BY run_id
) AS rst ON rst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(pst.*)::phase_status_timestamps[] AS plan_status_timestamps
    FROM phase_status_timestamps pst
    WHERE pst.phase = 'plan'
    GROUP BY run_id
) AS pst ON pst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(ast.*)::phase_status_timestamps[] AS apply_status_timestamps
    FROM phase_status_timestamps ast
    WHERE ast.phase = 'apply'
    GROUP BY run_id
) AS ast ON ast.run_id = runs.run_id
WHERE runs.run_id = sqlc.arg('run_id')
GROUP BY runs.run_id
FOR UPDATE OF runs, plans, applies
;

-- name: PutLockFile :one
UPDATE runs
SET lock_file = sqlc.arg('lock_file')
WHERE run_id = sqlc.arg('run_id')
RETURNING run_id
;

-- name: GetLockFileByID :one
SELECT lock_file
FROM runs
WHERE run_id = sqlc.arg('run_id')
;

-- name: UpdateRunStatus :one
UPDATE runs
SET
    status = sqlc.arg('status')
WHERE run_id = sqlc.arg('id')
RETURNING run_id
;

-- name: UpdateCancelSignaledAt :one
UPDATE runs
SET
    cancel_signaled_at = sqlc.arg('cancel_signaled_at')
WHERE run_id = sqlc.arg('id')
RETURNING run_id
;

-- name: DeleteRunByID :one
DELETE
FROM runs
WHERE run_id = sqlc.arg('run_id')
RETURNING run_id
;
