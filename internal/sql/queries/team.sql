-- name: InsertTeam :exec
INSERT INTO teams (
    team_id,
    name,
    created_at,
    organization_name,
    visibility,
    sso_team_id,
    permission_manage_workspaces,
    permission_manage_vcs,
    permission_manage_modules,
    permission_manage_providers,
    permission_manage_policies,
    permission_manage_policy_overrides
) VALUES (
    pggen.arg('id'),
    pggen.arg('name'),
    pggen.arg('created_at'),
    pggen.arg('organization_name'),
    pggen.arg('visibility'),
    pggen.arg('sso_team_id'),
    pggen.arg('permission_manage_workspaces'),
    pggen.arg('permission_manage_vcs'),
    pggen.arg('permission_manage_modules'),
    pggen.arg('permission_manage_providers'),
    pggen.arg('permission_manage_policies'),
    pggen.arg('permission_manage_policy_overrides')
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

-- name: FindTeamByTokenID :one
SELECT t.*
FROM teams t
JOIN team_tokens tt USING (team_id)
WHERE tt.team_id = pggen.arg('team_id')
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
    name = pggen.arg('name'),
    visibility = pggen.arg('visibility'),
    sso_team_id = pggen.arg('sso_team_id'),
    permission_manage_workspaces = pggen.arg('permission_manage_workspaces'),
    permission_manage_vcs = pggen.arg('permission_manage_vcs'),
    permission_manage_modules = pggen.arg('permission_manage_modules'),
    permission_manage_providers = pggen.arg('permission_manage_providers'),
    permission_manage_policies = pggen.arg('permission_manage_policies'),
    permission_manage_policy_overrides = pggen.arg('permission_manage_policy_overrides')
WHERE team_id = pggen.arg('team_id')
RETURNING team_id;

-- name: DeleteTeamByID :one
DELETE
FROM teams
WHERE team_id = pggen.arg('team_id')
RETURNING team_id
;
