-- name: InsertToken :exec
INSERT INTO tokens (
    token_id,
    token,
    created_at,
    description,
    user_id
) VALUES (
    pggen.arg('TokenID'),
    pggen.arg('Token'),
    pggen.arg('CreatedAt'),
    pggen.arg('Description'),
    pggen.arg('UserID')
);

-- name: DeleteTokenByID :one
DELETE
FROM tokens
WHERE token_id = pggen.arg('token_id')
RETURNING token_id
;
