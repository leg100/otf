-- name: InsertUser :exec
INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('updated_at'),
    pggen.arg('username')
);

-- name: FindUsers :many
SELECT u.*,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
;

-- name: FindUsersByOrganization :many
SELECT u.*,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
WHERE t.organization_name = pggen.arg('organization_name')
;

-- name: FindUsersByTeamID :many
SELECT
    u.*,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
WHERE t.team_id = pggen.arg('team_id')
;

-- name: FindUserByID :one
SELECT u.*,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
WHERE u.user_id = pggen.arg('user_id')
;

-- name: FindUserByUsername :one
SELECT u.*,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
WHERE u.username = pggen.arg('username')
;

-- name: FindUserByAuthenticationTokenID :one
SELECT u.*,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
    ) AS teams
FROM users u
JOIN tokens t ON u.username = t.username
WHERE t.token_id = pggen.arg('token_id')
;

-- name: UpdateUserSiteAdmins :many
UPDATE users
SET site_admin = true
WHERE username = ANY(pggen.arg('usernames'))
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
WHERE user_id = pggen.arg('user_id')
RETURNING user_id
;

-- name: DeleteUserByUsername :one
DELETE
FROM users
WHERE username = pggen.arg('username')
RETURNING user_id
;
