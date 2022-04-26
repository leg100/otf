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
    NOW(),
    NOW(),
    pggen.arg('Flash'),
    pggen.arg('Address'),
    pggen.arg('Expiry'),
    pggen.arg('UserID')
)
RETURNING *;

-- name: UpdateSessionUserID :one
UPDATE sessions
SET
    user_id = pggen.arg('user_id'),
    updated_at = NOW()
WHERE token = pggen.arg('token')
RETURNING *;

-- name: UpdateSessionExpiry :one
UPDATE sessions
SET
    expiry = pggen.arg('expiry'),
    updated_at = NOW()
WHERE token = pggen.arg('token')
RETURNING *;

-- name: UpdateSessionFlash :one
UPDATE sessions
SET
    flash = pggen.arg('flash'),
    updated_at = NOW()
WHERE token = pggen.arg('token')
RETURNING *;

-- name: DeleteSessionByToken :exec
DELETE
FROM sessions
WHERE token = pggen.arg('token');
