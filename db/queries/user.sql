-- name: InsertUser :one
INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username,
    current_organization
) VALUES (
    pggen.arg('ID'),
    NOW(),
    NOW(),
    pggen.arg('Username'),
    pggen.arg('CurrentOrganization')
)
RETURNING *;

-- name: FindUsers :many
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organization_memberships) AS organization_memberships
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN organization_memberships USING(user_id)
GROUP BY users.user_id
LIMIT pggen.arg('limit') OFFSET pggen.arg('offset')
;

-- name: FindUserByID :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organization_memberships) AS organization_memberships
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN organization_memberships USING(user_id)
WHERE users.user_id = pggen.arg('user_id')
GROUP BY users.user_id
;

-- name: FindUserByUsername :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organization_memberships) AS organization_memberships
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN organization_memberships USING(user_id)
WHERE users.username = pggen.arg('username')
GROUP BY users.user_id
;

-- name: FindUserBySessionToken :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organization_memberships) AS organization_memberships
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN organization_memberships USING(user_id)
WHERE sessions.token = pggen.arg('token')
GROUP BY users.user_id
;

-- name: FindUserByAuthenticationToken :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organization_memberships) AS organization_memberships
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN organization_memberships USING(user_id)
WHERE tokens.token = pggen.arg('token')
GROUP BY users.user_id
;

-- name: FindUserByAuthenticationTokenID :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organization_memberships) AS organization_memberships
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN organizations(JOIN organization_memberships USING(user_id)) USING 
WHERE tokens.token_id = pggen.arg('token_id')
GROUP BY users.user_id
;

-- name: UpdateUserCurrentOrganization :one
UPDATE users
SET
    current_organization = pggen.arg('current_organization'),
    updated_at = NOW()
WHERE user_id = pggen.arg('id')
RETURNING *;

-- name: DeleteUserByID :exec
DELETE
FROM users
WHERE user_id = pggen.arg('user_id');

-- name: DeleteUserByUsername :exec
DELETE
FROM users
WHERE username = pggen.arg('username');
