-- name: InsertUser :exec
INSERT INTO users (
    user_id,
    created_at,
    updated_at,
    username
) VALUES (
    pggen.arg('ID'),
    pggen.arg('CreatedAt'),
    pggen.arg('UpdatedAt'),
    pggen.arg('Username')
);

-- name: FindUsers :many
SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
GROUP BY u.user_id
;

-- name: FindUserByID :one
SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE u.user_id = pggen.arg('user_id')
GROUP BY u.user_id
;

-- name: FindUserByUsername :one
SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE u.username = pggen.arg('username')
GROUP BY u.user_id
;

-- name: FindUserBySessionToken :one
SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE s.token = pggen.arg('token')
GROUP BY u.user_id
;

-- name: FindUserByAuthenticationToken :one
SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE t.token = pggen.arg('token')
GROUP BY u.user_id
;

-- name: FindUserByAuthenticationTokenID :one
SELECT u.*,
    array_remove(array_agg(s), NULL) AS sessions,
    array_remove(array_agg(t), NULL) AS tokens,
    array_remove(array_agg(o), NULL) AS organizations
FROM users u
LEFT JOIN sessions s USING(user_id)
LEFT JOIN tokens t ON u.user_id = t.user_id
LEFT JOIN (organization_memberships om JOIN organizations o USING (organization_id)) ON u.user_id = om.user_id
WHERE t.token_id = pggen.arg('token_id')
GROUP BY u.user_id
;

-- name: DeleteUserByID :exec
DELETE
FROM users
WHERE user_id = pggen.arg('user_id');

-- name: DeleteUserByUsername :exec
DELETE
FROM users
WHERE username = pggen.arg('username');
