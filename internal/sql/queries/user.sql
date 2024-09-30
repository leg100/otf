-- name: InsertUser :exec
INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username
) VALUES (
    sqlc.arg('id'),
    sqlc.arg('created_at'),
    sqlc.arg('updated_at'),
    sqlc.arg('username')
);

-- name: FindUsers :many
SELECT
    u.*,
    array_agg(t.*)::"teams" AS teams
FROM users u
LEFT JOIN teams t USING (team_id)
;

-- name: FindUsersByOrganization :many
SELECT u.*,
    array_agg(t.*)::"teams" AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
WHERE t.organization_name = sqlc.arg('organization_name')
GROUP BY u.user_id
;

-- name: FindUsersByTeamID :many
SELECT
    u.*,
    array_agg(t.*)::"teams" AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
WHERE t.team_id = sqlc.arg('team_id')
GROUP BY u.user_id
;

-- name: FindUserByID :one
SELECT
    u.*,
    array_agg(t.*)::"teams" AS teams
FROM users u
LEFT JOIN teams t USING (team_id)
WHERE u.user_id = sqlc.arg('user_id')
GROUP BY u.user_id
;

-- name: FindUserByUsername :one
SELECT
    u.*,
    array_agg(t.*)::"teams" AS teams
FROM users u
LEFT JOIN teams t USING (team_id)
WHERE u.username = sqlc.arg('username')
GROUP BY u.user_id
;

-- name: FindUserByAuthenticationTokenID :one
SELECT
    u.*,
    array_agg(tt.*)::"teams" AS teams
FROM users u
JOIN tokens t ON u.username = t.username
LEFT JOIN teams tt USING (team_id)
WHERE t.token_id = sqlc.arg('token_id')
GROUP BY u.user_id
;

-- name: UpdateUserSiteAdmins :many
UPDATE users
SET site_admin = true
WHERE username = ANY(sqlc.arg('usernames')::text[])
RETURNING username
;

-- name: ResetUserSiteAdmins :many
UPDATE users
SET site_admin = false
WHERE site_admin = true
RETURNING username
;

-- name: DeleteUserByID :one
DELETE
FROM users
WHERE user_id = sqlc.arg('user_id')
RETURNING user_id
;

-- name: DeleteUserByUsername :one
DELETE
FROM users
WHERE username = sqlc.arg('username')
RETURNING user_id
;
