-- name: InsertRegistrySession :exec
INSERT INTO registry_sessions (
    token,
    expiry,
    organization_name
) VALUES (
    pggen.arg('token'),
    pggen.arg('expiry'),
    pggen.arg('organization_name')
);

-- name: FindRegistrySession :one
SELECT *
FROM registry_sessions
WHERE token = pggen.arg('token')
AND   expiry > current_timestamp
;

-- name: DeleteExpiredRegistrySessions :one
DELETE
FROM registry_sessions
WHERE expiry < current_timestamp
RETURNING token
;
