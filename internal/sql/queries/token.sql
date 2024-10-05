-- name: InsertToken :exec
INSERT INTO tokens (
    token_id,
    created_at,
    description,
    username
) VALUES (
    sqlc.arg('token_id'),
    sqlc.arg('created_at'),
    sqlc.arg('description'),
    sqlc.arg('username')
);

-- name: FindTokensByUsername :many
SELECT *
FROM tokens
WHERE username = sqlc.arg('username')
;

-- name: FindTokenByID :one
SELECT *
FROM tokens
WHERE token_id = sqlc.arg('token_id')
;

-- name: DeleteTokenByID :one
DELETE
FROM tokens
WHERE token_id = sqlc.arg('token_id')
RETURNING token_id
;
