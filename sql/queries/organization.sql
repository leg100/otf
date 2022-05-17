-- FindOrganizationByName finds an organization by name.
--
-- name: FindOrganizationByName :one
SELECT * FROM organizations WHERE name = pggen.arg('name');

-- name: FindOrganizationByNameForUpdate :one
SELECT *
FROM organizations
WHERE name = pggen.arg('name')
FOR UPDATE
;

-- name: FindOrganizations :many
SELECT *
FROM organizations
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset');

-- name: CountOrganizations :one
SELECT count(*)
FROM organizations;

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
    current_timestamp,
    current_timestamp,
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
    updated_at = current_timestamp
WHERE name = pggen.arg('name')
RETURNING *;

-- name: UpdateOrganizationSessionRememberByName :one
UPDATE organizations
SET
    session_remember = pggen.arg('session_remember'),
    updated_at = current_timestamp
WHERE name = pggen.arg('name')
RETURNING *;

-- name: UpdateOrganizationSessionTimeoutByName :one
UPDATE organizations
SET
    session_timeout = pggen.arg('session_timeout'),
    updated_at = current_timestamp
WHERE name = pggen.arg('name')
RETURNING *;

-- DeleteOrganization deletes an organization by id.
--
-- name: DeleteOrganization :exec
DELETE
FROM organizations
WHERE name = pggen.arg('name');
