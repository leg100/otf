-- name: InsertAgentToken :exec
INSERT INTO agent_tokens (
    agent_token_id,
    created_at,
    description,
    agent_pool_id
) VALUES (
    pggen.arg('agent_token_id'),
    pggen.arg('created_at'),
    pggen.arg('description'),
    pggen.arg('agent_pool_id')
);

-- name: FindAgentTokenByID :one
SELECT *
FROM agent_tokens
WHERE agent_token_id = pggen.arg('agent_token_id')
;

-- name: FindAgentTokensByAgentPoolID :many
SELECT *
FROM agent_tokens
WHERE agent_pool_id = pggen.arg('agent_pool_id')
ORDER BY created_at DESC
;

-- name: DeleteAgentTokenByID :one
DELETE
FROM agent_tokens
WHERE agent_token_id = pggen.arg('agent_token_id')
RETURNING agent_token_id
;
