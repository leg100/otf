-- name: InsertWorkspace :exec
INSERT INTO workspaces (
    workspace_id,
    created_at,
    updated_at,
    allow_destroy_plan,
    auto_apply,
    can_queue_destroy_plan,
    description,
    environment,
    execution_mode,
    file_triggers_enabled,
    global_remote_state,
    migration_environment,
    name,
    queue_all_runs,
    speculative_enabled,
    source_name,
    source_url,
    structured_run_output_enabled,
    terraform_version,
    trigger_prefixes,
    working_directory,
    organization_name
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('updated_at'),
    pggen.arg('allow_destroy_plan'),
    pggen.arg('auto_apply'),
    pggen.arg('can_queue_destroy_plan'),
    pggen.arg('description'),
    pggen.arg('environment'),
    pggen.arg('execution_mode'),
    pggen.arg('file_triggers_enabled'),
    pggen.arg('global_remote_state'),
    pggen.arg('migration_environment'),
    pggen.arg('name'),
    pggen.arg('queue_all_runs'),
    pggen.arg('speculative_enabled'),
    pggen.arg('source_name'),
    pggen.arg('source_url'),
    pggen.arg('structured_run_output_enabled'),
    pggen.arg('terraform_version'),
    pggen.arg('trigger_prefixes'),
    pggen.arg('working_directory'),
    pggen.arg('organization_name')
);

-- name: FindWorkspaces :many
SELECT
    w.*,
    (u.*)::"users" AS user_lock,
    (r.*)::"runs" AS run_lock,
    (vr.*)::"workspace_repos" AS workspace_repo,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
LEFT JOIN users u ON w.lock_user_id = u.user_id
LEFT JOIN runs r ON w.lock_run_id = r.run_id
LEFT JOIN (workspace_repos vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
WHERE w.name                LIKE pggen.arg('prefix') || '%'
AND   w.organization_name   LIKE ANY(pggen.arg('organization_names'))
ORDER BY w.updated_at DESC
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
;

-- name: CountWorkspaces :one
SELECT count(*)
FROM workspaces
WHERE name LIKE pggen.arg('prefix') || '%'
AND   organization_name LIKE ANY(pggen.arg('organization_names'))
;

-- name: FindWorkspacesByWebhookID :many
SELECT
    w.*,
    (ul.*)::"users" AS user_lock,
    (rl.*)::"runs" AS run_lock,
    (vr.*)::"workspace_repos" AS workspace_repo,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
LEFT JOIN users ul ON w.lock_user_id = ul.user_id
LEFT JOIN runs rl ON w.lock_run_id = rl.run_id
JOIN (workspace_repos vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
WHERE h.webhook_id = pggen.arg('webhook_id')
;

-- name: FindWorkspacesByUserID :many
SELECT
    w.*,
    (ul.*)::"users" AS user_lock,
    (rl.*)::"runs" AS run_lock,
    (vr.*)::"workspace_repos" AS workspace_repo,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
JOIN workspace_permissions p USING (workspace_id)
LEFT JOIN users ul ON w.lock_user_id = ul.user_id
LEFT JOIN runs rl ON w.lock_run_id = rl.run_id
LEFT JOIN (workspace_repos vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
JOIN teams t USING (team_id)
JOIN team_memberships tm USING (team_id)
JOIN users u ON tm.user_id = u.user_id
WHERE w.organization_name  = pggen.arg('organization_name')
AND   u.user_id            = pggen.arg('user_id')
ORDER BY w.updated_at DESC
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
;

-- name: CountWorkspacesByUserID :one
SELECT count(*)
FROM workspaces w
JOIN workspace_permissions p USING (workspace_id)
JOIN teams t USING (team_id)
JOIN team_memberships tm USING (team_id)
JOIN users u USING (user_id)
WHERE w.organization_name = pggen.arg('organization_name')
AND   u.user_id = pggen.arg('user_id')
;

-- name: FindWorkspaceIDByRunID :one
SELECT w.workspace_id
FROM workspaces w
JOIN runs r USING (workspace_id)
WHERE r.run_id = pggen.arg('run_id')
;

-- name: FindWorkspaceIDByStateVersionID :one
SELECT workspace_id
FROM state_versions
WHERE state_version_id = pggen.arg('state_version_id')
;

-- name: FindWorkspaceIDByCVID :one
SELECT workspace_id
FROM configuration_versions
WHERE configuration_version_id = pggen.arg('configuration_version_id')
;

-- name: FindWorkspaceIDByName :one
SELECT workspaces.workspace_id
FROM workspaces
WHERE workspaces.name = pggen.arg('name')
AND workspaces.organization_name = pggen.arg('organization_name');

-- name: FindWorkspaceByName :one
SELECT w.*,
    (u.*)::"users" AS user_lock,
    (r.*)::"runs" AS run_lock,
    (vr.*)::"workspace_repos" AS workspace_repo,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
LEFT JOIN users u ON w.lock_user_id = u.user_id
LEFT JOIN runs r ON w.lock_run_id = r.run_id
LEFT JOIN (workspace_repos vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
WHERE w.name              = pggen.arg('name')
AND   w.organization_name = pggen.arg('organization_name')
;

-- name: FindWorkspaceByID :one
SELECT w.*,
    (u.*)::"users" AS user_lock,
    (r.*)::"runs" AS run_lock,
    (vr.*)::"workspace_repos" AS workspace_repo,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
LEFT JOIN users u ON w.lock_user_id = u.user_id
LEFT JOIN runs r ON w.lock_run_id = r.run_id
LEFT JOIN (workspace_repos vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
WHERE w.workspace_id = pggen.arg('id')
;

-- name: FindWorkspaceByIDForUpdate :one
SELECT w.*,
    (u.*)::"users" AS user_lock,
    (r.*)::"runs" AS run_lock,
    (vr.*)::"workspace_repos" AS workspace_repo,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
LEFT JOIN users u ON w.lock_user_id = u.user_id
LEFT JOIN runs r ON w.lock_run_id = r.run_id
LEFT JOIN (workspace_repos vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
WHERE w.workspace_id = pggen.arg('id')
FOR UPDATE OF w;

-- name: UpdateWorkspaceByID :one
UPDATE workspaces
SET
    allow_destroy_plan              = pggen.arg('allow_destroy_plan'),
    auto_apply                      = pggen.arg('auto_apply'),
    description                     = pggen.arg('description'),
    execution_mode                  = pggen.arg('execution_mode'),
    name                            = pggen.arg('name'),
    queue_all_runs                  = pggen.arg('queue_all_runs'),
    speculative_enabled             = pggen.arg('speculative_enabled'),
    structured_run_output_enabled   = pggen.arg('structured_run_output_enabled'),
    terraform_version               = pggen.arg('terraform_version'),
    trigger_prefixes                = pggen.arg('trigger_prefixes'),
    working_directory               = pggen.arg('working_directory'),
    updated_at                      = pggen.arg('updated_at')
WHERE workspace_id = pggen.arg('id')
RETURNING workspace_id;

-- name: UpdateWorkspaceLockByID :exec
UPDATE workspaces
SET
    lock_user_id = pggen.arg('user_id'),
    lock_run_id = pggen.arg('run_id')
WHERE workspace_id = pggen.arg('workspace_id');

-- name: UpdateWorkspaceLatestRun :exec
UPDATE workspaces
SET latest_run_id = pggen.arg('run_id')
WHERE workspace_id = pggen.arg('workspace_id');

-- name: DeleteWorkspaceByID :exec
DELETE
FROM workspaces
WHERE workspace_id = pggen.arg('workspace_id');
