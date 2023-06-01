-- name: InsertWorkspace :exec
INSERT INTO workspaces (
    workspace_id,
    created_at,
    updated_at,
    allow_destroy_plan,
    auto_apply,
    branch,
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
    pggen.arg('branch'),
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
    (
        SELECT array_agg(name)
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    (
        SELECT (u.*)::"users"
        FROM users u
        WHERE u.username = w.lock_username
    ) AS user_lock,
    (
        SELECT (rl.*)::"runs"
        FROM runs rl
        WHERE rl.run_id = w.lock_run_id
    ) AS run_lock,
    (
        SELECT (rc.*)::"repo_connections"
        FROM repo_connections rc
        WHERE rc.workspace_id = w.workspace_id
    ) AS workspace_connection,
    (
        SELECT (wh.*)::"webhooks"
        FROM webhooks wh
        JOIN repo_connections rc USING (webhook_id)
        WHERE rc.workspace_id = w.workspace_id
    ) AS webhook
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN (workspace_tags wt JOIN tags t USING (tag_id)) ON wt.workspace_id = w.workspace_id
WHERE w.name                LIKE '%' || pggen.arg('search') || '%'
AND   w.organization_name   LIKE ANY(pggen.arg('organization_names'))
GROUP BY w.workspace_id, r.status
HAVING array_agg(t.name) @> pggen.arg('tags')
ORDER BY w.updated_at DESC
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
;

-- name: CountWorkspaces :one
SELECT count(distinct(w.workspace_id))
FROM workspaces w
LEFT JOIN (workspace_tags wt JOIN tags t USING (tag_id)) ON w.workspace_id = wt.workspace_id
WHERE w.name              LIKE '%' || pggen.arg('search') || '%'
AND   w.organization_name LIKE ANY(pggen.arg('organization_names'))
AND   CASE WHEN cardinality(pggen.arg('tags')::text[]) > 0 THEN t.name LIKE ANY(pggen.arg('tags'))
      ELSE 1 = 1
      END
;

-- name: FindWorkspacesByWebhookID :many
SELECT
    w.*,
    (
        SELECT array_agg(name)
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    (ul.*)::"users" AS user_lock,
    (rl.*)::"runs" AS run_lock,
    (vr.*)::"repo_connections" AS workspace_connection,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
LEFT JOIN users ul ON w.lock_username = ul.username
LEFT JOIN runs rl ON w.lock_run_id = rl.run_id
LEFT JOIN runs r ON w.latest_run_id = r.run_id
JOIN (repo_connections vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
WHERE h.webhook_id = pggen.arg('webhook_id')
;

-- name: FindWorkspacesByUsername :many
SELECT
    w.*,
    (
        SELECT array_agg(name)
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    (ul.*)::"users" AS user_lock,
    (rl.*)::"runs" AS run_lock,
    (vr.*)::"repo_connections" AS workspace_connection,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
JOIN workspace_permissions p USING (workspace_id)
LEFT JOIN users ul ON w.lock_username = ul.username
LEFT JOIN runs rl ON w.lock_run_id = rl.run_id
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN (repo_connections vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
JOIN teams t USING (team_id)
JOIN team_memberships tm USING (team_id)
JOIN users u ON tm.username = u.username
WHERE w.organization_name  = pggen.arg('organization_name')
AND   u.username           = pggen.arg('username')
ORDER BY w.updated_at DESC
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
;

-- name: CountWorkspacesByUsername :one
SELECT count(*)
FROM workspaces w
JOIN workspace_permissions p USING (workspace_id)
JOIN teams t USING (team_id)
JOIN team_memberships tm USING (team_id)
JOIN users u USING (username)
WHERE w.organization_name = pggen.arg('organization_name')
AND   u.username          = pggen.arg('username')
;

-- name: FindWorkspaceByName :one
SELECT w.*,
    (
        SELECT array_agg(name)
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    (ul.*)::"users" AS user_lock,
    (rl.*)::"runs" AS run_lock,
    (vr.*)::"repo_connections" AS workspace_connection,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
LEFT JOIN users ul ON w.lock_username = ul.username
LEFT JOIN runs rl ON w.lock_run_id = rl.run_id
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN (repo_connections vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
WHERE w.name              = pggen.arg('name')
AND   w.organization_name = pggen.arg('organization_name')
;

-- name: FindWorkspaceByID :one
SELECT w.*,
    (
        SELECT array_agg(name)
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    (ul.*)::"users" AS user_lock,
    (rl.*)::"runs" AS run_lock,
    (vr.*)::"repo_connections" AS workspace_connection,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
LEFT JOIN users ul ON w.lock_username = ul.username
LEFT JOIN runs rl ON w.lock_run_id = rl.run_id
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN (repo_connections vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
WHERE w.workspace_id = pggen.arg('id')
;

-- name: FindWorkspaceByIDForUpdate :one
SELECT w.*,
    (
        SELECT array_agg(name)
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    (ul.*)::"users" AS user_lock,
    (rl.*)::"runs" AS run_lock,
    (vr.*)::"repo_connections" AS workspace_connection,
    (h.*)::"webhooks" AS webhook
FROM workspaces w
LEFT JOIN users ul ON w.lock_username = ul.username
LEFT JOIN runs rl ON w.lock_run_id = rl.run_id
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN (repo_connections vr JOIN webhooks h USING (webhook_id)) ON w.workspace_id = vr.workspace_id
WHERE w.workspace_id = pggen.arg('id')
FOR UPDATE OF w;

-- name: UpdateWorkspaceByID :one
UPDATE workspaces
SET
    allow_destroy_plan              = pggen.arg('allow_destroy_plan'),
    auto_apply                      = pggen.arg('auto_apply'),
    branch                          = pggen.arg('branch'),
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
    lock_username = pggen.arg('username'),
    lock_run_id = pggen.arg('run_id')
WHERE workspace_id = pggen.arg('workspace_id');

-- name: UpdateWorkspaceLatestRun :exec
UPDATE workspaces
SET latest_run_id = pggen.arg('run_id')
WHERE workspace_id = pggen.arg('workspace_id');

-- name: UpdateWorkspaceCurrentStateVersionID :one
UPDATE workspaces
SET current_state_version_id = pggen.arg('state_version_id')
WHERE workspace_id = pggen.arg('workspace_id')
RETURNING workspace_id;

-- name: DeleteWorkspaceByID :exec
DELETE
FROM workspaces
WHERE workspace_id = pggen.arg('workspace_id');
