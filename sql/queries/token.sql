-- name: InsertToken :exec
INSERT INTO tokens (
    token_id,
    token,
    created_at,
    description,
    username
) VALUES (
    pggen.arg('token_id'),
    pggen.arg('token'),
    pggen.arg('created_at'),
    pggen.arg('description'),
    pggen.arg('username')
);

-- name: FindTokensByUsername :many
SELECT *
FROM tokens
WHERE username = pggen.arg('username')
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
