-- name: InsertWorkspace :exec
INSERT INTO workspaces (
    workspace_id,
    created_at,
    updated_at,
    agent_pool_id,
    allow_cli_apply,
    allow_destroy_plan,
    auto_apply,
    branch,
    can_queue_destroy_plan,
    description,
    environment,
    execution_mode,
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
    trigger_patterns,
    vcs_tags_regex,
    working_directory,
    organization_name
) VALUES (
    sqlc.arg('id'),
    sqlc.arg('created_at'),
    sqlc.arg('updated_at'),
    sqlc.arg('agent_pool_id'),
    sqlc.arg('allow_cli_apply'),
    sqlc.arg('allow_destroy_plan'),
    sqlc.arg('auto_apply'),
    sqlc.arg('branch'),
    sqlc.arg('can_queue_destroy_plan'),
    sqlc.arg('description'),
    sqlc.arg('environment'),
    sqlc.arg('execution_mode'),
    sqlc.arg('global_remote_state'),
    sqlc.arg('migration_environment'),
    sqlc.arg('name'),
    sqlc.arg('queue_all_runs'),
    sqlc.arg('speculative_enabled'),
    sqlc.arg('source_name'),
    sqlc.arg('source_url'),
    sqlc.arg('structured_run_output_enabled'),
    sqlc.arg('terraform_version'),
    sqlc.arg('trigger_prefixes'),
    sqlc.arg('trigger_patterns'),
    sqlc.arg('vcs_tags_regex'),
    sqlc.arg('working_directory'),
    sqlc.arg('organization_name')
);

-- name: FindWorkspaces :many
SELECT
    w.*,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
        GROUP BY wt.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
LEFT JOIN (workspace_tags wt JOIN tags t USING (tag_id)) ON wt.workspace_id = w.workspace_id
WHERE w.name                LIKE '%' || sqlc.arg('search') || '%'
AND   w.organization_name   LIKE ANY(sqlc.arg('organization_names')::text[])
GROUP BY w.workspace_id, r.status, rc.vcs_provider_id, rc.repo_path
HAVING array_agg(t.name) @> sqlc.arg('tags')::text[]
ORDER BY w.updated_at DESC
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int
;

-- name: CountWorkspaces :one
WITH
    workspaces AS (
        SELECT w.workspace_id
        FROM workspaces w
        LEFT JOIN (workspace_tags wt JOIN tags t USING (tag_id)) ON w.workspace_id = wt.workspace_id
        WHERE w.name              LIKE '%' || sqlc.arg('search') || '%'
        AND   w.organization_name LIKE ANY(sqlc.arg('organization_names')::text[])
        GROUP BY w.workspace_id
        HAVING array_agg(t.name) @> sqlc.arg('tags')::text[]
    )
SELECT count(*)
FROM workspaces
;

-- name: FindWorkspacesByConnection :many
SELECT
    w.*,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN users ul ON w.lock_username = ul.username
LEFT JOIN runs rl ON w.lock_run_id = rl.run_id
LEFT JOIN runs r ON w.latest_run_id = r.run_id
JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
WHERE rc.vcs_provider_id = sqlc.arg('vcs_provider_id')
AND   rc.repo_path = sqlc.arg('repo_path')
;

-- name: FindWorkspacesByUsername :many
SELECT
    w.*,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
JOIN workspace_permissions p USING (workspace_id)
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
JOIN teams t USING (team_id)
JOIN team_memberships tm USING (team_id)
JOIN users u ON tm.username = u.username
WHERE w.organization_name  = sqlc.arg('organization_name')
AND   u.username           = sqlc.arg('username')
ORDER BY w.updated_at DESC
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int
;

-- name: CountWorkspacesByUsername :one
SELECT count(*)
FROM workspaces w
JOIN workspace_permissions p USING (workspace_id)
JOIN teams t USING (team_id)
JOIN team_memberships tm USING (team_id)
JOIN users u USING (username)
WHERE w.organization_name = sqlc.arg('organization_name')
AND   u.username          = sqlc.arg('username')
;

-- name: FindWorkspaceByName :one
SELECT
    w.*,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
LEFT JOIN (workspace_tags wt JOIN tags t USING (tag_id)) ON w.workspace_id = rc.workspace_id
WHERE w.name              = sqlc.arg('name')
AND   w.organization_name = sqlc.arg('organization_name')
;

-- name: FindWorkspaceByID :one
SELECT
    w.*,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN users ul ON w.lock_username = ul.username
LEFT JOIN runs rl ON w.lock_run_id = rl.run_id
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
WHERE w.workspace_id = sqlc.arg('id')
;

-- name: FindWorkspaceByIDForUpdate :one
SELECT
    w.*,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
WHERE w.workspace_id = sqlc.arg('id')
FOR UPDATE OF w;

-- name: UpdateWorkspaceByID :one
UPDATE workspaces
SET
    agent_pool_id                 = sqlc.arg('agent_pool_id'),
    allow_destroy_plan            = sqlc.arg('allow_destroy_plan'),
    allow_cli_apply               = sqlc.arg('allow_cli_apply'),
    auto_apply                    = sqlc.arg('auto_apply'),
    branch                        = sqlc.arg('branch'),
    description                   = sqlc.arg('description'),
    execution_mode                = sqlc.arg('execution_mode'),
    global_remote_state           = sqlc.arg('global_remote_state'),
    name                          = sqlc.arg('name'),
    queue_all_runs                = sqlc.arg('queue_all_runs'),
    speculative_enabled           = sqlc.arg('speculative_enabled'),
    structured_run_output_enabled = sqlc.arg('structured_run_output_enabled'),
    terraform_version             = sqlc.arg('terraform_version'),
    trigger_prefixes              = sqlc.arg('trigger_prefixes'),
    trigger_patterns              = sqlc.arg('trigger_patterns'),
    vcs_tags_regex                = sqlc.arg('vcs_tags_regex'),
    working_directory             = sqlc.arg('working_directory'),
    updated_at                    = sqlc.arg('updated_at')
WHERE workspace_id = sqlc.arg('id')
RETURNING workspace_id;

-- name: UpdateWorkspaceLockByID :exec
UPDATE workspaces
SET
    lock_username = sqlc.arg('username'),
    lock_run_id = sqlc.arg('run_id')
WHERE workspace_id = sqlc.arg('workspace_id');

-- name: UpdateWorkspaceLatestRun :exec
UPDATE workspaces
SET latest_run_id = sqlc.arg('run_id')
WHERE workspace_id = sqlc.arg('workspace_id');

-- name: UpdateWorkspaceCurrentStateVersionID :one
UPDATE workspaces
SET current_state_version_id = sqlc.arg('state_version_id')
WHERE workspace_id = sqlc.arg('workspace_id')
RETURNING workspace_id;

-- name: DeleteWorkspaceByID :exec
DELETE
FROM workspaces
WHERE workspace_id = sqlc.arg('workspace_id');
