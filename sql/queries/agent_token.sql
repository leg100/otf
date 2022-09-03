-- name: InsertAgentToken :exec
INSERT INTO agent_tokens (
    token_id,
    token,
    created_at,
    description,
    organization_name
) VALUES (
    pggen.arg('token_id'),
    pggen.arg('token'),
    pggen.arg('created_at'),
    pggen.arg('description'),
    pggen.arg('organization_name')
);

-- name: FindAgentToken :one
SELECT *
FROM agent_tokens
WHERE token = pggen.arg('token')
;

-- name: FindAgentTokens :many
SELECT *
FROM agent_tokens
WHERE organization_name = pggen.arg('organization_name')
ORDER BY created_at DESC
;

-- name: DeleteAgentTokenByID :one
DELETE
FROM agent_tokens
WHERE token_id = pggen.arg('token_id')
RETURNING token_id
;
