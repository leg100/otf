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

-- name: FindAllAgents :many
SELECT *
FROM agents;

-- name: FindAgents :many
SELECT *
FROM agents

-- name: DeleteAgent :one
DELETE
FROM agents
WHERE agent_id = pggen.arg('agent_id')
RETURNING *;
