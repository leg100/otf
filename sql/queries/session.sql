-- name: InsertSession :exec
INSERT INTO sessions (
    token,
    created_at,
    address,
    expiry,
    username
) VALUES (
    pggen.arg('token'),
    pggen.arg('created_at'),
    pggen.arg('address'),
    pggen.arg('expiry'),
    pggen.arg('username')
);

-- name: FindSessionsByUsername :many
SELECT *
FROM sessions
WHERE username = pggen.arg('username')
AND   expiry > current_timestamp
;

-- name: FindSessionByToken :one
SELECT *
FROM sessions
WHERE token = pggen.arg('token')
;

-- name: UpdateSessionExpiry :one
UPDATE sessions
SET
    expiry = pggen.arg('expiry')
WHERE token = pggen.arg('token')
RETURNING token
;

-- name: DeleteSessionByToken :one
DELETE
FROM sessions
WHERE token = pggen.arg('token')
RETURNING token
;

-- name: DeleteSessionsExpired :one
DELETE
FROM sessions
WHERE expiry < current_timestamp
RETURNING token
;
