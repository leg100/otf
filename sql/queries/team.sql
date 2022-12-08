-- name: InsertTeam :exec
INSERT INTO teams (
    team_id,
    name,
    created_at,
    organization_id
) VALUES (
    pggen.arg('id'),
    pggen.arg('name'),
    pggen.arg('created_at'),
    pggen.arg('organization_id')
);

-- name: FindTeamsByOrg :many
SELECT
    t.*,
    (o.*)::"organizations" AS organization
FROM teams t
JOIN organizations o USING (organization_id)
WHERE o.name = pggen.arg('organization_name')
;

-- name: FindTeamByName :one
SELECT
    t.*,
    (o.*)::"organizations" AS organization
FROM teams t
JOIN organizations o USING (organization_id)
WHERE t.name = pggen.arg('name')
AND   o.name = pggen.arg('organization_name')
;

-- name: FindTeamByNameForUpdate :one
SELECT
    t.*,
    (o.*)::"organizations" AS organization
FROM teams t
JOIN organizations o USING (organization_id)
WHERE t.name = pggen.arg('name')
AND   o.name = pggen.arg('organization_name')
FOR UPDATE OF t
;

-- name: UpdateTeamByName :one
UPDATE teams
SET
    permission_manage_workspaces = pggen.arg('permission_manage_workspaces'),
    permission_manage_vcs = pggen.arg('permission_manage_vcs')
FROM organizations o
WHERE teams.organization_id = o.organization_id
AND   o.name = pggen.arg('organization_name')
AND   teams.name = pggen.arg('name')
RETURNING team_id;

-- name: DeleteTeamByName :one
DELETE
FROM teams
USING organizations
WHERE teams.organization_id = organizations.organization_id
AND   teams.name = pggen.arg('name')
AND   organizations.name = pggen.arg('organization_name')
RETURNING team_id
;

