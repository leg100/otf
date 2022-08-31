-- name: InsertAgentToken :exec
INSERT INTO agent_tokens (
    token_id,
    token,
    created_at,
    organization_id
) VALUES (
    pggen.arg('TokenID'),
    pggen.arg('Token'),
    pggen.arg('CreatedAt'),
    pggen.arg('OrganizationID')
);

-- name: FindAgentToken :one
SELECT *
FROM agent_tokens
WHERE token = pggen.arg('token')
;

-- name: DeleteAgentTokenByID :one
DELETE
FROM agent_tokens
WHERE token_id = pggen.arg('token_id')
RETURNING token_id
;
