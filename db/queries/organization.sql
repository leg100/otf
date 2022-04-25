-- FindOrganizationByName finds an organization by name.
--
-- name: FindOrganizationByName :one
SELECT * FROM organizations WHERE name = pggen.arg('name');

-- InsertOrganization inserts an organization and returns the entire row.
--
-- name: InsertOrganization :one
INSERT INTO organizations (
    organization_id,
    created_at,
    updated_at,
    name,
    session_remember,
    session_timeout
) VALUES (
    pggen.arg('ID'),
    NOW(),
    NOW(),
    pggen.arg('Name'),
    pggen.arg('SessionRemember'),
    pggen.arg('SessionTimeout')
)
RETURNING *;

-- UpdateOrganizationNameByName updates an organization with a new name,
-- identifying the organization with its existing name, and returns the
-- updated row.
--
-- name: UpdateOrganizationNameByName :one
UPDATE organizations
SET
    name = pggen.arg('new_name'),
    updated_at = NOW()
WHERE name = pggen.arg('name')
RETURNING *;

-- DeleteOrganization deletes an organization by id.
--
-- name: DeleteOrganization :exec
DELETE
FROM organizations
WHERE name = pggen.arg('name');
