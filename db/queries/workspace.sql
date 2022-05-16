-- InsertWorkspace inserts a workspace and returns the entire row.
--
-- name: InsertWorkspace :one
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
    locked,
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
    current_timestamp,
    current_timestamp,
    pggen.arg('AllowDestroyPlan'),
    pggen.arg('AutoApply'),
    pggen.arg('CanQueueDestroyPlan'),
    pggen.arg('Description'),
    pggen.arg('Environment'),
    pggen.arg('ExecutionMode'),
    pggen.arg('FileTriggersEnabled'),
    pggen.arg('GlobalRemoteState'),
    pggen.arg('Locked'),
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
)
RETURNING *;

-- name: FindWorkspaces :many
SELECT
    workspaces.*,
    (organizations.*)::"organizations" AS organization
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name LIKE pggen.arg('prefix') || '%'
AND organizations.name = pggen.arg('organization_name')
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
;

-- name: CountWorkspaces :one
SELECT count(*)
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name LIKE pggen.arg('prefix') || '%'
AND organizations.name = pggen.arg('organization_name')
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
SELECT workspaces.*, (organizations.*)::"organizations" AS organization
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name = pggen.arg('name')
AND organizations.name = pggen.arg('organization_name');

-- name: FindWorkspaceByNameForUpdate :one
SELECT workspaces.*, (organizations.*)::"organizations" AS organization
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name = pggen.arg('name')
AND organizations.name = pggen.arg('organization_name')
FOR UPDATE;

-- FindWorkspaceByID finds a workspace by id.
--
-- name: FindWorkspaceByID :one
SELECT workspaces.*, (organizations.*)::"organizations" AS organization
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.workspace_id = pggen.arg('id');

-- name: FindWorkspaceByIDForUpdate :one
SELECT workspaces.*, (organizations.*)::"organizations" AS organization
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.workspace_id = pggen.arg('id')
FOR UPDATE;

-- UpdateWorkspaceNameByID updates an workspace with a new name,
-- identifying the workspace with its id, and returns the
-- updated row.
--
-- name: UpdateWorkspaceNameByID :one
UPDATE workspaces
SET
    name = pggen.arg('name'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

-- UpdateWorkspaceAllowDestroyPlanByID updates the AllowDestroyPlan
-- attribute on a workspace identified by id, and returns the updated row.
--
-- name: UpdateWorkspaceAllowDestroyPlanByID :one
UPDATE workspaces
SET
    allow_destroy_plan = pggen.arg('allow_destroy_plan'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

-- name: UpdateWorkspaceExecutionModeByID :one
UPDATE workspaces
SET
    execution_mode = pggen.arg('execution_mode'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

-- name: UpdateWorkspaceLockByID :one
UPDATE workspaces
SET
    locked = pggen.arg('lock'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

-- name: UpdateWorkspaceDescriptionByID :one
UPDATE workspaces
SET
    description = pggen.arg('description'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

-- name: UpdateWorkspaceSpeculativeEnabledByID :one
UPDATE workspaces
SET
    speculative_enabled = pggen.arg('speculative_enabled'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

-- name: UpdateWorkspaceStructuredRunOutputEnabledByID :one
UPDATE workspaces
SET
    structured_run_output_enabled = pggen.arg('structured_run_output_enabled'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

-- name: UpdateWorkspaceTerraformVersionByID :one
UPDATE workspaces
SET
    terraform_version = pggen.arg('terraform_version'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

-- name: UpdateWorkspaceTriggerPrefixesByID :one
UPDATE workspaces
SET
    trigger_prefixes = pggen.arg('trigger_prefixes'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

-- name: UpdateWorkspaceWorkingDirectoryByID :one
UPDATE workspaces
SET
    working_directory = pggen.arg('working_directory'),
    updated_at = current_timestamp
WHERE workspace_id = pggen.arg('id')
RETURNING updated_at;

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
