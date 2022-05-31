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

-- name: UpdateSessionUserID :one
UPDATE sessions
SET
    user_id = pggen.arg('user_id')
WHERE token = pggen.arg('token')
RETURNING token;

-- name: UpdateSessionExpiry :one
UPDATE sessions
SET
    expiry = pggen.arg('expiry')
WHERE token = pggen.arg('token')
RETURNING token;

-- name: DeleteSessionByToken :exec
DELETE
FROM sessions
WHERE token = pggen.arg('token');

-- name: DeleteSessionsExpired :exec
DELETE
FROM sessions
WHERE expiry < current_timestamp;
