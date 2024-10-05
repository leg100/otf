-- name: InsertAgent :exec
INSERT INTO agents (
    agent_id,
    name,
    version,
    max_jobs,
    ip_address,
    last_ping_at,
    last_status_at,
    status,
    agent_pool_id
) VALUES (
    sqlc.arg('agent_id'),
    sqlc.arg('name'),
    sqlc.arg('version'),
    sqlc.arg('max_jobs'),
    sqlc.arg('ip_address'),
    sqlc.arg('last_ping_at'),
    sqlc.arg('last_status_at'),
    sqlc.arg('status'),
    sqlc.arg('agent_pool_id')
);

-- name: UpdateAgent :one
UPDATE agents
SET status = sqlc.arg('status'),
    last_ping_at = sqlc.arg('last_ping_at'),
    last_status_at = sqlc.arg('last_status_at')
WHERE agent_id = sqlc.arg('agent_id')
RETURNING *;

-- name: FindAgents :many
SELECT
    a.*,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
GROUP BY a.agent_id
ORDER BY a.last_ping_at DESC;

-- name: FindAgentsByOrganization :many
SELECT
    a.*,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.organization_name = sqlc.arg('organization_name')
GROUP BY a.agent_id
ORDER BY last_ping_at DESC;

-- name: FindAgentsByPoolID :many
SELECT
    a.*,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.agent_pool_id = sqlc.arg('agent_pool_id')
GROUP BY a.agent_id
ORDER BY last_ping_at DESC;

-- name: FindServerAgents :many
SELECT
    a.*,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
WHERE agent_pool_id IS NULL
GROUP BY a.agent_id
ORDER BY last_ping_at DESC;

-- name: FindAgentByID :one
SELECT
    a.*,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
LEFT JOIN jobs j USING (agent_id)
WHERE a.agent_id = sqlc.arg('agent_id')
GROUP BY a.agent_id;

-- name: FindAgentByIDForUpdate :one
SELECT
    a.*,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.agent_id = j.agent_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM agents a
WHERE a.agent_id = sqlc.arg('agent_id')
FOR UPDATE OF a;

-- name: DeleteAgent :one
DELETE
FROM agents
WHERE agent_id = sqlc.arg('agent_id')
RETURNING *;
