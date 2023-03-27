-- name: InsertToken :exec
INSERT INTO tokens (
    token_id,
    token,
    created_at,
    description,
    user_id
) VALUES (
    pggen.arg('token_id'),
    pggen.arg('token'),
    pggen.arg('created_at'),
    pggen.arg('description'),
    pggen.arg('user_id')
);

-- name: FindTokensByUserID :many
SELECT *
FROM tokens
WHERE user_id = pggen.arg('user_id')
;

-- name: FindTokenByID :one
SELECT *
FROM tokens
WHERE token_id = pggen.arg('token_id')
;

-- name: DeleteTokenByID :one
DELETE
FROM tokens
WHERE token_id = pggen.arg('token_id')
RETURNING token_id
;
