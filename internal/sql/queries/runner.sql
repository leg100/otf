-- name: InsertRunner :exec
INSERT INTO runners (
    runner_id,
    name,
    version,
    max_jobs,
    ip_address,
    last_ping_at,
    last_status_at,
    status,
    agent_pool_id
) VALUES (
    sqlc.arg('runner_id'),
    sqlc.arg('name'),
    sqlc.arg('version'),
    sqlc.arg('max_jobs'),
    sqlc.arg('ip_address'),
    sqlc.arg('last_ping_at'),
    sqlc.arg('last_status_at'),
    sqlc.arg('status'),
    sqlc.arg('agent_pool_id')
);

-- name: UpdateRunner :one
UPDATE runners
SET status = sqlc.arg('status'),
    last_ping_at = sqlc.arg('last_ping_at'),
    last_status_at = sqlc.arg('last_status_at')
WHERE runner_id = sqlc.arg('runner_id')
RETURNING *;

-- name: FindRunners :many
SELECT
    a.*,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
LEFT JOIN agent_pools ap USING (agent_pool_id)
WHERE (
	ap.organization_name = sqlc.arg('organization_name')
	OR sqlc.arg('organization_name') IS NULL
	OR ap.agent_pool_id IS NULL
)
AND   (ap.agent_pool_id = sqlc.arg('agent_pool_id')::text OR sqlc.arg('agent_pool_id')::text IS NULL)
AND   (
	(sqlc.arg('is_server')::bool AND ap.agent_pool_id IS NULL)
	OR (NOT sqlc.arg('is_server')::bool AND ap.agent_pool_id IS NOT NULL)
	OR sqlc.arg('is_server')::bool IS NULL
)
ORDER BY last_ping_at DESC
;

-- name: CountRunners :one
SELECT count(a.*)
FROM runners a
LEFT JOIN agent_pools ap USING (agent_pool_id)
WHERE (
	ap.organization_name = sqlc.arg('organization_name')
	OR sqlc.arg('organization_name') IS NULL
	OR ap.agent_pool_id IS NULL
)
AND   (ap.agent_pool_id = sqlc.arg('agent_pool_id')::text OR sqlc.arg('agent_pool_id')::text IS NULL)
AND   (
	(sqlc.arg('is_server')::bool AND ap.agent_pool_id IS NULL)
	OR (NOT sqlc.arg('is_server')::bool AND ap.agent_pool_id IS NOT NULL)
	OR sqlc.arg('is_server')::bool IS NULL
)
;

-- name: FindRunnerByID :one
SELECT
    a.*,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
LEFT JOIN agent_pools ap USING (agent_pool_id)
LEFT JOIN jobs j USING (runner_id)
WHERE a.runner_id = sqlc.arg('runner_id');

-- name: FindRunnerByIDForUpdate :one
SELECT
    a.*,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
LEFT JOIN agent_pools ap USING (agent_pool_id)
WHERE a.runner_id = sqlc.arg('runner_id')
FOR UPDATE OF a;

-- name: DeleteRunner :one
DELETE
FROM runners
WHERE runner_id = sqlc.arg('runner_id')
RETURNING *;
