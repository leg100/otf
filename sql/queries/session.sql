-- name: InsertSession :one
INSERT INTO sessions (
    token,
    created_at,
    updated_at,
    flash,
    address,
    expiry,
    user_id
) VALUES (
    pggen.arg('Token'),
    current_timestamp,
    current_timestamp,
    pggen.arg('Flash'),
    pggen.arg('Address'),
    pggen.arg('Expiry'),
    pggen.arg('UserID')
)
RETURNING *;

-- name: FindSessionFlashByToken :one
SELECT flash
FROM sessions
WHERE token = pggen.arg('token');

-- name: UpdateSessionFlashByToken :exec
UPDATE sessions
SET
    flash = pggen.arg('flash')
WHERE token = pggen.arg('token');

-- name: UpdateSessionUserID :one
UPDATE sessions
SET
    user_id = pggen.arg('user_id'),
    updated_at = current_timestamp
WHERE token = pggen.arg('token')
RETURNING *;

-- name: UpdateSessionExpiry :one
UPDATE sessions
SET
    expiry = pggen.arg('expiry'),
    updated_at = current_timestamp
WHERE token = pggen.arg('token')
RETURNING *;

-- name: UpdateSessionFlash :one
UPDATE sessions
SET
    flash = pggen.arg('flash'),
    updated_at = current_timestamp
WHERE token = pggen.arg('token')
RETURNING *;

-- name: DeleteSessionByToken :exec
DELETE
FROM sessions
WHERE token = pggen.arg('token');

-- name: DeleteSessionsExpired :exec
DELETE
FROM sessions
WHERE expiry < current_timestamp;
