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
    organization_id
) VALUES (
    pggen.arg('ID'),
    pggen.arg('CreatedAt'),
    pggen.arg('UpdatedAt'),
    pggen.arg('AllowDestroyPlan'),
    pggen.arg('AutoApply'),
    pggen.arg('CanQueueDestroyPlan'),
    pggen.arg('Description'),
    pggen.arg('Environment'),
    pggen.arg('ExecutionMode'),
    pggen.arg('FileTriggersEnabled'),
    pggen.arg('GlobalRemoteState'),
    pggen.arg('MigrationEnvironment'),
    pggen.arg('Name'),
    pggen.arg('QueueAllRuns'),
    pggen.arg('SpeculativeEnabled'),
    pggen.arg('SourceName'),
    pggen.arg('SourceUrl'),
    pggen.arg('StructuredRunOutputEnabled'),
    pggen.arg('TerraformVersion'),
    pggen.arg('TriggerPrefixes'),
    pggen.arg('WorkingDirectory'),
    pggen.arg('OrganizationID')
);

-- name: FindWorkspaces :many
SELECT
    w.*,
    o.name AS organization_name,
    (u.*)::"users" AS user_lock,
    (r.*)::"runs" AS run_lock,
    CASE WHEN pggen.arg('include_organization') THEN (o.*)::"organizations" END AS organization
FROM workspaces w
JOIN organizations o USING (organization_id)
LEFT JOIN users u ON w.lock_user_id = u.user_id
LEFT JOIN runs r ON w.lock_run_id = r.run_id
WHERE w.name LIKE pggen.arg('prefix') || '%'
AND   o.name LIKE ANY(pggen.arg('organization_names'))
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
;

-- name: CountWorkspaces :one
SELECT count(*)
FROM workspaces w
JOIN organizations o USING (organization_id)
WHERE w.name LIKE pggen.arg('prefix') || '%'
AND   o.name LIKE ANY(pggen.arg('organization_names'))
;

-- name: FindWorkspaceIDByName :one
SELECT workspaces.workspace_id
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name = pggen.arg('name')
AND organizations.name = pggen.arg('organization_name');

-- FindWorkspaceByName finds a workspace by name and organization name.
--
-- name: FindWorkspaceByName :one
SELECT w.*,
    organizations.name AS organization_name,
    (u.*)::"users" AS user_lock,
    (r.*)::"runs" AS run_lock,
    CASE WHEN pggen.arg('include_organization') THEN (organizations.*)::"organizations" END AS organization
FROM workspaces w
JOIN organizations USING (organization_id)
LEFT JOIN users u ON w.lock_user_id = u.user_id
LEFT JOIN runs r ON w.lock_run_id = r.run_id
WHERE w.name = pggen.arg('name')
AND organizations.name = pggen.arg('organization_name');

-- name: FindWorkspaceByID :one
SELECT w.*,
    organizations.name AS organization_name,
    (u.*)::"users" AS user_lock,
    (r.*)::"runs" AS run_lock,
    CASE WHEN pggen.arg('include_organization') THEN (organizations.*)::"organizations" END AS organization
FROM workspaces w
JOIN organizations USING (organization_id)
LEFT JOIN users u ON w.lock_user_id = u.user_id
LEFT JOIN runs r ON w.lock_run_id = r.run_id
WHERE w.workspace_id = pggen.arg('id');

-- name: FindWorkspaceByIDForUpdate :one
SELECT w.*,
    organizations.name AS organization_name,
    (u.*)::"users" AS user_lock,
    (r.*)::"runs" AS run_lock,
    NULL::"organizations" AS organization
FROM workspaces w
JOIN organizations USING (organization_id)
LEFT JOIN users u ON w.lock_user_id = u.user_id
LEFT JOIN runs r ON w.lock_run_id = r.run_id
WHERE w.workspace_id = pggen.arg('id')
FOR UPDATE OF w;

-- name: UpdateWorkspaceByID :one
UPDATE workspaces
SET
    allow_destroy_plan = pggen.arg('allow_destroy_plan'),
    description = pggen.arg('description'),
    execution_mode = pggen.arg('execution_mode'),
    name = pggen.arg('name'),
    queue_all_runs = pggen.arg('queue_all_runs'),
    speculative_enabled = pggen.arg('speculative_enabled'),
    structured_run_output_enabled = pggen.arg('structured_run_output_enabled'),
    terraform_version = pggen.arg('terraform_version'),
    trigger_prefixes = pggen.arg('trigger_prefixes'),
    working_directory = pggen.arg('working_directory'),
    updated_at = pggen.arg('updated_at')
WHERE workspace_id = pggen.arg('id')
RETURNING workspace_id;

-- name: UpdateWorkspaceLockByID :exec
UPDATE workspaces
SET
    lock_user_id = pggen.arg('user_id'),
    lock_run_id = pggen.arg('run_id')
WHERE workspace_id = pggen.arg('workspace_id');

-- DeleteOrganization deletes an organization by id.
-- DeleteWorkspaceByID deletes a workspace by id.
--
-- name: DeleteWorkspaceByID :exec
DELETE
FROM workspaces
WHERE workspace_id = pggen.arg('workspace_id');

-- DeleteWorkspaceByName deletes a workspace by name and organization name.
--
-- name: DeleteWorkspaceByName :exec
DELETE
FROM workspaces
USING organizations
WHERE workspaces.organization_id = organizations.organization_id
AND workspaces.name = pggen.arg('name')
AND organizations.name = pggen.arg('organization_name');
