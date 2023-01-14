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
    array_remove(array_agg(o.name), NULL) AS organizations,
    array_remove(array_agg(teams), NULL) AS teams
FROM users u
LEFT JOIN (organization_memberships om JOIN organizations o ON om.organization_name = o.name) ON u.user_id = om.user_id
LEFT JOIN (team_memberships tm JOIN teams USING (team_id)) ON u.user_id = tm.user_id
GROUP BY u.user_id
;

-- name: FindUsersByOrganization :many
SELECT u.*,
    (
        SELECT array_remove(array_agg(o.name), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.user_id = u.user_id
    ) AS organizations,
    array_remove(array_agg(teams), NULL) AS teams
FROM users u
JOIN (organization_memberships om JOIN organizations o ON om.organization_name = o.name) ON u.user_id = om.user_id
LEFT JOIN (team_memberships tm JOIN teams USING (team_id)) ON u.user_id = tm.user_id
WHERE o.name = pggen.arg('organization_name')
GROUP BY u.user_id
;

-- name: FindUsersByTeam :many
SELECT
    u.*,
    array_remove(array_agg(o.name), NULL) AS organizations,
    array_remove(array_agg(t), NULL) AS teams
FROM users u
JOIN team_memberships tm USING (user_id)
JOIN teams t USING (team_id)
JOIN organizations o ON o.name = t.organization_name
WHERE o.name = pggen.arg('organization_name')
AND   t.name = pggen.arg('team_name')
GROUP BY u.user_id
;

-- name: FindUsersByTeamID :many
SELECT
    u.*,
    (
        SELECT array_agg(o.name)
        FROM organizations o
        JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.user_id = u.user_id
    ) AS organizations,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.user_id = u.user_id
    ) AS teams
FROM users u
JOIN team_memberships tm USING (user_id)
JOIN teams t USING (team_id)
WHERE t.team_id = pggen.arg('team_id')
GROUP BY u.user_id
;

-- name: FindUserByID :one
SELECT u.*,
    (
        SELECT array_remove(array_agg(o.name), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.user_id = u.user_id
    ) AS organizations,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM teams t
        LEFT JOIN team_memberships tm USING (team_id)
        WHERE tm.user_id = u.user_id
    ) AS teams
FROM users u
WHERE u.user_id = pggen.arg('user_id')
GROUP BY u.user_id
;

-- name: FindUserByUsername :one
SELECT u.*,
    (
        SELECT array_remove(array_agg(o.name), NULL)
        FROM organizations o
        LEFT JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.user_id = u.user_id
    ) AS organizations,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM teams t
        LEFT JOIN team_memberships tm USING (team_id)
        WHERE tm.user_id = u.user_id
    ) AS teams
FROM users u
WHERE u.username = pggen.arg('username')
GROUP BY u.user_id
;

-- name: FindUserBySessionToken :one
SELECT u.*,
    (
        SELECT array_agg(o.name)
        FROM organizations o
        JOIN organization_memberships om ON om.organization_name = o.name
        WHERE om.user_id = u.user_id
    ) AS organizations,
    (
        SELECT array_agg(t)
        FROM teams t
        JOIN team_memberships tm USING (team_id)
        WHERE tm.user_id = u.user_id
    ) AS teams
FROM users u
JOIN sessions s ON u.user_id = s.user_id AND s.expiry > current_timestamp
WHERE s.token = pggen.arg('token')
GROUP BY u.user_id
;

-- name: FindUserByAuthenticationToken :one
SELECT u.*,
    (
        select array_remove(array_agg(o.name), null)
        from organizations o
        left join organization_memberships om ON om.organization_name = o.name
        where om.user_id = u.user_id
    ) as organizations,
    (
        SELECT array_remove(array_agg(t), NULL)
        FROM teams t
        LEFT JOIN team_memberships tm USING (team_id)
        WHERE tm.user_id = u.user_id
    ) AS teams
FROM users u
LEFT JOIN tokens t ON u.user_id = t.user_id
WHERE t.token = pggen.arg('token')
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
