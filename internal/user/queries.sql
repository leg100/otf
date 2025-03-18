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
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
;

-- name: FindUsersByOrganization :many
SELECT
    u.*,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
WHERE t.organization_name = sqlc.arg('organization_name')
GROUP BY u.user_id
;

-- name: FindUsersByTeamID :many
SELECT
    u.*,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
JOIN team_memberships tm USING (username)
JOIN teams t USING (team_id)
WHERE t.team_id = sqlc.arg('team_id')
GROUP BY u.user_id
;

-- name: FindUserByID :one
SELECT
    u.*,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
WHERE u.user_id = sqlc.arg('user_id')
;

-- name: FindUserByUsername :one
SELECT
    u.*,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
WHERE u.username = sqlc.arg('username')
;

-- name: FindUserByAuthenticationTokenID :one
SELECT
    u.*,
    (
        SELECT array_agg(t.*)::teams[]
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.username = u.username
        GROUP BY tm.username
    ) AS teams
FROM users u
JOIN tokens t ON u.username = t.username
WHERE t.token_id = sqlc.arg('token_id')
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


--
-- user tokens
--

-- name: InsertToken :exec
INSERT INTO tokens (
    token_id,
    created_at,
    description,
    username
) VALUES (
    sqlc.arg('token_id'),
    sqlc.arg('created_at'),
    sqlc.arg('description'),
    sqlc.arg('username')
);

-- name: FindTokensByUsername :many
SELECT *
FROM tokens
WHERE username = sqlc.arg('username')
;

-- name: FindTokenByID :one
SELECT *
FROM tokens
WHERE token_id = sqlc.arg('token_id')
;

-- name: DeleteTokenByID :one
DELETE
FROM tokens
WHERE token_id = sqlc.arg('token_id')
RETURNING token_id
;
