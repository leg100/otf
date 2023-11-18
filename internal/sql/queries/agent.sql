-- name: InsertAgent :exec
INSERT INTO agents (
    agent_id,
    name,
    concurrency,
    server,
    ip_address,
    last_ping_at,
    last_status_at,
    status,
    agent_pool_id
) VALUES (
    pggen.arg('agent_id'),
    pggen.arg('name'),
    pggen.arg('concurrency'),
    pggen.arg('server'),
    pggen.arg('ip_address'),
    pggen.arg('last_ping_at'),
    pggen.arg('last_status_at'),
    pggen.arg('status'),
    pggen.arg('agent_pool_id')
);

-- name: UpdateAgent :one
UPDATE agents
SET status = pggen.arg('status'),
    last_ping_at = pggen.arg('last_ping_at'),
    last_status_at = pggen.arg('last_status_at')
WHERE agent_id = pggen.arg('agent_id')
RETURNING *;

-- name: FindAgents :many
SELECT *
FROM agents;

-- name: FindAgentsByOrganization :many
SELECT a.*
FROM agents a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.organization_name = pggen.arg('organization_name');

-- name: FindAgentsByPoolID :many
SELECT a.*
FROM agents a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.agent_pool_id = pggen.arg('agent_pool_id');

-- name: FindServerAgents :many
SELECT *
FROM agents
WHERE server
ORDER BY last_ping_at DESC;

-- name: FindAgentByID :one
SELECT *
FROM agents
WHERE agent_id = pggen.arg('agent_id');

-- name: FindAgentByIDForUpdate :one
SELECT *
FROM agents
WHERE agent_id = pggen.arg('agent_id')
FOR UPDATE;

-- name: DeleteAgent :one
DELETE
FROM agents
WHERE agent_id = pggen.arg('agent_id')
RETURNING *;
