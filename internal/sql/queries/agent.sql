-- name: InsertAgent :exec
INSERT INTO agents (
    agent_id,
    name,
    concurrency,
    server,
    ip_address,
    last_ping_at,
    status,
    agent_token_id
) VALUES (
    pggen.arg('agent_id'),
    pggen.arg('name'),
    pggen.arg('concurrency'),
    pggen.arg('server'),
    pggen.arg('ip_address'),
    current_timestamp,
    pggen.arg('status'),
    pggen.arg('agent_token_id')
);

-- name: UpdateAgentStatus :one
UPDATE agents
SET status = pggen.arg('status'), last_ping_at = current_timestamp
WHERE agent_id = pggen.arg('agent_id')
RETURNING *;

-- name: FindAgents :many
SELECT *
FROM agents;

-- name: FindAgentsByOrganization :many
SELECT a.*
FROM agents a
JOIN (agent_tokens at JOIN agent_pools ap USING (agent_pool_id)) USING (agent_token_id)
WHERE ap.organization_name = pggen.arg('organization_name');

-- name: FindServerAgents :many
SELECT *
FROM agents
WHERE server;

-- name: DeleteAgent :one
DELETE
FROM agents
WHERE agent_id = pggen.arg('agent_id')
RETURNING *;
