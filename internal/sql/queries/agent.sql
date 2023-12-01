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
    pggen.arg('agent_id'),
    pggen.arg('name'),
    pggen.arg('version'),
    pggen.arg('max_jobs'),
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
WHERE ap.organization_name = pggen.arg('organization_name')
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
WHERE ap.agent_pool_id = pggen.arg('agent_pool_id')
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
WHERE a.agent_id = pggen.arg('agent_id')
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
WHERE agent_id = pggen.arg('agent_id')
FOR UPDATE OF a;

-- name: DeleteAgent :one
DELETE
FROM agents
WHERE agent_id = pggen.arg('agent_id')
RETURNING *;
