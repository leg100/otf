-- name: InsertUser :one
INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username,
    current_organization
) VALUES (
    pggen.arg('ID'),
    current_timestamp,
    current_timestamp,
    pggen.arg('Username'),
    pggen.arg('CurrentOrganization')
)
RETURNING *;

-- name: FindUsers :many
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organizations) AS organizations
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN (organization_memberships JOIN organizations USING (organization_id)) USING(user_id)
GROUP BY users.user_id
;

-- name: FindUserByID :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organizations) AS organizations
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN (organization_memberships JOIN organizations USING (organization_id)) USING(user_id)
WHERE users.user_id = pggen.arg('user_id')
GROUP BY users.user_id
;

-- name: FindUserByUsername :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organizations) AS organizations
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN (organization_memberships JOIN organizations USING (organization_id)) USING(user_id)
WHERE users.username = pggen.arg('username')
AND sessions.expiry > current_timestamp
GROUP BY users.user_id
;

-- name: FindUserBySessionToken :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organizations) AS organizations
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN (organization_memberships JOIN organizations USING (organization_id)) USING(user_id)
WHERE sessions.token = pggen.arg('token')
AND sessions.expiry > current_timestamp
GROUP BY users.user_id
;

-- name: FindUserByAuthenticationToken :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organizations) AS organizations
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN (organization_memberships JOIN organizations USING (organization_id)) USING(user_id)
WHERE tokens.token = pggen.arg('token')
AND sessions.expiry > current_timestamp
GROUP BY users.user_id
;

-- name: FindUserByAuthenticationTokenID :one
SELECT users.*,
    array_agg(sessions) AS sessions,
    array_agg(tokens) AS tokens,
    array_agg(organizations) AS organizations
FROM users
JOIN sessions USING(user_id)
JOIN tokens USING(user_id)
JOIN (organization_memberships JOIN organizations USING (organization_id)) USING(user_id)
WHERE tokens.token_id = pggen.arg('token_id')
AND sessions.expiry > current_timestamp
GROUP BY users.user_id
;

-- name: UpdateUserCurrentOrganization :one
UPDATE users
SET
    current_organization = pggen.arg('current_organization'),
    updated_at = current_timestamp
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
