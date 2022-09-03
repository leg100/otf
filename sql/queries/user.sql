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
    (
        SELECT array_remove(array_agg(s), NULL)
        FROM sessions s
        WHERE s.user_id = u.user_id
        AND s.expiry > current_timestamp
    ) AS sessions,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM tokens t
        WHERE t.user_id = u.user_id
    ) AS tokens,
    (
        SELECT array_remove(array_agg(o), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om USING (organization_id)
        WHERE om.user_id = u.user_id
    ) AS organizations
FROM users u
WHERE u.user_id = pggen.arg('user_id')
GROUP BY u.user_id
;

-- name: FindUserByUsername :one
SELECT u.*,
    (
        SELECT array_remove(array_agg(s), NULL)
        FROM sessions s
        WHERE s.user_id = u.user_id
        AND s.expiry > current_timestamp
    ) AS sessions,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM tokens t
        WHERE t.user_id = u.user_id
    ) AS tokens,
    (
        SELECT array_remove(array_agg(o), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om USING (organization_id)
        WHERE om.user_id = u.user_id
    ) AS organizations
FROM users u
WHERE u.username = pggen.arg('username')
GROUP BY u.user_id
;

-- name: FindUserBySessionToken :one
SELECT u.*,
    (
        SELECT array_remove(array_agg(s), NULL)
        FROM sessions s
        WHERE s.user_id = u.user_id
        AND s.expiry > current_timestamp
    ) AS sessions,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM tokens t
        WHERE t.user_id = u.user_id
    ) AS tokens,
    (
        SELECT array_remove(array_agg(o), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om USING (organization_id)
        WHERE om.user_id = u.user_id
    ) AS organizations
FROM users u
JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
WHERE s.token = pggen.arg('token')
GROUP BY u.user_id
;

-- name: FindUserByAuthenticationToken :one
SELECT u.*,
    (
        select array_remove(array_agg(s), null)
        from sessions s
        where s.user_id = u.user_id
        and s.expiry > current_timestamp
    ) as sessions,
    (
        select array_remove(array_agg(t), null)
        from tokens t
        where t.user_id = u.user_id
    ) as tokens,
    (
        select array_remove(array_agg(o), null)
        from organizations o
        left join organization_memberships om using (organization_id)
        where om.user_id = u.user_id
    ) as organizations
FROM users u
LEFT JOIN tokens t ON u.user_id = t.user_id
WHERE t.token = pggen.arg('token')
GROUP BY u.user_id
;

-- name: FindUserByAuthenticationTokenID :one
SELECT u.*,
    (
        SELECT array_remove(array_agg(s), NULL)
        FROM sessions s
        WHERE s.user_id = u.user_id
        AND s.expiry > current_timestamp
    ) AS sessions,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM tokens t
        WHERE t.user_id = u.user_id
    ) AS tokens,
    (
        SELECT array_remove(array_agg(o), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om USING (organization_id)
        WHERE om.user_id = u.user_id
    ) AS organizations
FROM users u
JOIN tokens t ON u.user_id = t.user_id
WHERE t.token_id = pggen.arg('token_id')
GROUP BY u.user_id
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
