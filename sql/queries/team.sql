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

-- name: FindTeamByID :one
SELECT
    t.*,
    (o.*)::"organizations" AS organization
FROM teams t
JOIN organizations o USING (organization_id)
WHERE t.team_id = pggen.arg('team_id')
;

-- name: FindTeamByIDForUpdate :one
SELECT
    t.*,
    (o.*)::"organizations" AS organization
FROM teams t
JOIN organizations o USING (organization_id)
WHERE t.team_id = pggen.arg('team_id')
FOR UPDATE OF t
;

-- name: UpdateTeamByID :one
UPDATE teams
SET
    permission_manage_workspaces = pggen.arg('permission_manage_workspaces'),
    permission_manage_vcs = pggen.arg('permission_manage_vcs')
WHERE team_id = pggen.arg('team_id')
RETURNING team_id;

-- name: DeleteTeamByID :one
DELETE
FROM teams
WHERE team_id = pggen.arg('team_id')
RETURNING team_id
;

