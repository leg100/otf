-- name: InsertToken :one
INSERT INTO tokens (
    token_id,
    token,
    created_at,
    updated_at,
    description,
    user_id
) VALUES (
    pggen.arg('TokenID'),
    pggen.arg('Token'),
    NOW(),
    NOW(),
    pggen.arg('Description'),
    pggen.arg('UserID')
)
RETURNING *;

-- name: DeleteTokenByID :exec
DELETE
FROM tokens
WHERE token_id = pggen.arg('token_id');
