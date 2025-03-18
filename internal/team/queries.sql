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
    sqlc.arg('id'),
    sqlc.arg('name'),
    sqlc.arg('created_at'),
    sqlc.arg('organization_name'),
    sqlc.arg('visibility'),
    sqlc.arg('sso_team_id'),
    sqlc.arg('permission_manage_workspaces'),
    sqlc.arg('permission_manage_vcs'),
    sqlc.arg('permission_manage_modules'),
    sqlc.arg('permission_manage_providers'),
    sqlc.arg('permission_manage_policies'),
    sqlc.arg('permission_manage_policy_overrides')
);

-- name: FindTeamsByOrg :many
SELECT *
FROM teams
WHERE organization_name = sqlc.arg('organization_name')
;

-- name: FindTeamByName :one
SELECT *
FROM teams
WHERE name              = sqlc.arg('name')
AND   organization_name = sqlc.arg('organization_name')
;

-- name: FindTeamByID :one
SELECT *
FROM teams
WHERE team_id = sqlc.arg('team_id')
;

-- name: FindTeamByTokenID :one
SELECT t.*
FROM teams t
JOIN team_tokens tt USING (team_id)
WHERE tt.team_token_id = sqlc.arg('token_id')
;

-- name: FindTeamByIDForUpdate :one
SELECT *
FROM teams t
WHERE team_id = sqlc.arg('team_id')
FOR UPDATE OF t
;

-- name: UpdateTeamByID :one
UPDATE teams
SET
    name = sqlc.arg('name'),
    visibility = sqlc.arg('visibility'),
    sso_team_id = sqlc.arg('sso_team_id'),
    permission_manage_workspaces = sqlc.arg('permission_manage_workspaces'),
    permission_manage_vcs = sqlc.arg('permission_manage_vcs'),
    permission_manage_modules = sqlc.arg('permission_manage_modules'),
    permission_manage_providers = sqlc.arg('permission_manage_providers'),
    permission_manage_policies = sqlc.arg('permission_manage_policies'),
    permission_manage_policy_overrides = sqlc.arg('permission_manage_policy_overrides')
WHERE team_id = sqlc.arg('team_id')
RETURNING team_id;

-- name: DeleteTeamByID :one
DELETE
FROM teams
WHERE team_id = sqlc.arg('team_id')
RETURNING team_id
;

--
-- team tokens
--

-- name: InsertTeamToken :exec
INSERT INTO team_tokens (
    team_token_id,
    created_at,
    team_id,
    expiry
) VALUES (
    sqlc.arg('team_token_id'),
    sqlc.arg('created_at'),
    sqlc.arg('team_id'),
    sqlc.arg('expiry')
) ON CONFLICT (team_id) DO UPDATE
  SET team_token_id = sqlc.arg('team_token_id'),
      created_at    = sqlc.arg('created_at'),
      expiry        = sqlc.arg('expiry');

-- name: FindTeamTokensByID :many
SELECT *
FROM team_tokens
WHERE team_id = sqlc.arg('team_id')
;

-- name: DeleteTeamTokenByID :one
DELETE
FROM team_tokens
WHERE team_id = sqlc.arg('team_id')
RETURNING team_token_id
;
