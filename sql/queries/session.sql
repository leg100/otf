-- name: InsertSession :exec
INSERT INTO sessions (
    token,
    created_at,
    address,
    expiry,
    user_id
) VALUES (
    pggen.arg('Token'),
    pggen.arg('CreatedAt'),
    pggen.arg('Address'),
    pggen.arg('Expiry'),
    pggen.arg('UserID')
);

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
