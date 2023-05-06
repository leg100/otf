-- name: InsertTeam :exec
INSERT INTO teams (
    team_id,
    name,
    created_at,
    organization_name
) VALUES (
    pggen.arg('id'),
    pggen.arg('name'),
    pggen.arg('created_at'),
    pggen.arg('organization_name')
);

-- name: FindTeamsByOrg :many
SELECT *
FROM teams
WHERE organization_name = pggen.arg('organization_name')
;

-- name: FindTeamByName :one
SELECT *
FROM teams
WHERE name              = pggen.arg('name')
AND   organization_name = pggen.arg('organization_name')
;

-- name: FindTeamByID :one
SELECT *
FROM teams
WHERE team_id = pggen.arg('team_id')
;

-- name: FindTeamByIDForUpdate :one
SELECT *
FROM teams t
WHERE team_id = pggen.arg('team_id')
FOR UPDATE OF t
;

-- name: UpdateTeamByID :one
UPDATE teams
SET
    permission_manage_workspaces = pggen.arg('permission_manage_workspaces'),
    permission_manage_vcs = pggen.arg('permission_manage_vcs'),
    permission_manage_registry = pggen.arg('permission_manage_registry')
WHERE team_id = pggen.arg('team_id')
RETURNING team_id;

-- name: DeleteTeamByID :one
DELETE
FROM teams
WHERE team_id = pggen.arg('team_id')
RETURNING team_id
;
