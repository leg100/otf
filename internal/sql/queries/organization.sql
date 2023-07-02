-- name: InsertOrganization :exec
INSERT INTO organizations (
    organization_id,
    created_at,
    updated_at,
    name,
    email,
    collaborator_auth_policy,
    session_remember,
    session_timeout,
    allow_force_delete_workspaces
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('updated_at'),
    pggen.arg('name'),
    pggen.arg('email'),
    pggen.arg('collaborator_auth_policy'),
    pggen.arg('session_remember'),
    pggen.arg('session_timeout'),
    pggen.arg('allow_force_delete_workspaces')
);

-- name: FindOrganizationNameByWorkspaceID :one
SELECT organization_name
FROM workspaces
WHERE workspace_id = pggen.arg('workspace_id')
;

-- name: FindOrganizationByName :one
SELECT * FROM organizations WHERE name = pggen.arg('name');

-- name: FindOrganizationByID :one
SELECT * FROM organizations WHERE organization_id = pggen.arg('organization_id');

-- name: FindOrganizationByNameForUpdate :one
SELECT *
FROM organizations
WHERE name = pggen.arg('name')
FOR UPDATE
;

-- name: FindOrganizations :many
SELECT *
FROM organizations
WHERE name LIKE ANY(pggen.arg('names'))
ORDER BY updated_at DESC
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: CountOrganizations :one
SELECT count(*)
FROM organizations
WHERE name LIKE ANY(pggen.arg('names'))
;

-- name: UpdateOrganizationByName :one
UPDATE organizations
SET
    name = pggen.arg('new_name'),
    email = pggen.arg('email'),
    collaborator_auth_policy = pggen.arg('collaborator_auth_policy'),
    session_remember = pggen.arg('session_remember'),
    session_timeout = pggen.arg('session_timeout'),
    allow_force_delete_workspaces = pggen.arg('allow_force_delete_workspaces'),
    updated_at = pggen.arg('updated_at')
WHERE name = pggen.arg('name')
RETURNING organization_id;

-- name: DeleteOrganizationByName :one
DELETE
FROM organizations
WHERE name = pggen.arg('name')
RETURNING organization_id;
