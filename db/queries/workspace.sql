-- FindWorkspaceByName finds a workspace by name and organization name.
--
-- name: FindWorkspaceByName :one
SELECT workspaces.*, (organizations.*)::"organizations" AS organization
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name = pggen.arg('name')
AND organizations.name = pggen.arg('organization_name');

-- FindWorkspaceByID finds a workspace by id.
--
-- name: FindWorkspaceByID :one
SELECT workspaces.*, (organizations.*)::"organizations" AS organization
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.workspace_id = pggen.arg('id');

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
