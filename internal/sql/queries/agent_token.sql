-- name: InsertAgentToken :exec
INSERT INTO agent_tokens (
    agent_token_id,
    created_at,
    description,
    organization_name
) VALUES (
    pggen.arg('agent_token_id'),
    pggen.arg('created_at'),
    pggen.arg('description'),
    pggen.arg('organization_name')
);

-- name: FindAgentTokenByID :one
SELECT *
FROM agent_tokens
WHERE agent_token_id = pggen.arg('agent_token_id')
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
WHERE agent_token_id = pggen.arg('agent_token_id')
RETURNING agent_token_id
;
