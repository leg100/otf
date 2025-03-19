--
-- runners
--

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
ORDER BY a.last_ping_at DESC;

-- name: FindRunnersByOrganization :many
SELECT
    a.*,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.organization_name = sqlc.arg('organization_name')
ORDER BY last_ping_at DESC;

-- name: FindRunnersByPoolID :many
SELECT
    a.*,
    ap::"agent_pools" AS agent_pool,
    ( SELECT count(*)
      FROM jobs j
      WHERE a.runner_id = j.runner_id
      AND j.status IN ('allocated', 'running')
    ) AS current_jobs
FROM runners a
JOIN agent_pools ap USING (agent_pool_id)
WHERE ap.agent_pool_id = sqlc.arg('agent_pool_id')
ORDER BY last_ping_at DESC;

-- name: FindServerRunners :many
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
WHERE agent_pool_id IS NULL
ORDER BY last_ping_at DESC;

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

--
-- jobs
--

-- name: InsertJob :exec
INSERT INTO jobs (
    job_id,
    run_id,
    phase,
    status
) VALUES (
    sqlc.arg('job_id'),
    sqlc.arg('run_id'),
    sqlc.arg('phase'),
    sqlc.arg('status')
);

-- name: FindJobs :many
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
;

-- name: FindJob :one
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.job_id = sqlc.arg('job_id')
;

-- name: FindJobForUpdate :one
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.job_id = sqlc.arg('job_id')
FOR UPDATE OF j
;


-- FindUnfinishedJobForUpdateByRunID finds an unfinished job belonging to a run.
-- (There should only be one such job for a run).
--
-- name: FindUnfinishedJobForUpdateByRunID :one
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.run_id = sqlc.arg('run_id')
AND   j.status IN ('unallocated', 'allocated', 'running')
FOR UPDATE OF j
;

-- name: FindAllocatedJobs :many
SELECT
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.runner_id = sqlc.arg('runner_id')
AND   j.status = 'allocated';

-- Find signaled jobs and then immediately update signal with null.
--
-- name: FindAndUpdateSignaledJobs :many
UPDATE jobs AS j
SET signaled = NULL
FROM runs r, workspaces w
WHERE j.run_id = r.run_id
AND   r.workspace_id = w.workspace_id
AND   j.runner_id = sqlc.arg('runner_id')
AND   j.status = 'running'
AND   j.signaled IS NOT NULL
RETURNING
    j.job_id,
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.runner_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
;

-- name: UpdateJob :one
UPDATE jobs
SET status   = sqlc.arg('status'),
    signaled = sqlc.arg('signaled'),
    runner_id = sqlc.arg('runner_id')
WHERE job_id = sqlc.arg('job_id')
RETURNING *;


--
-- agent pools
--

-- name: InsertAgentPool :exec
INSERT INTO agent_pools (
    agent_pool_id,
    name,
    created_at,
    organization_name,
    organization_scoped
) VALUES (
    sqlc.arg('agent_pool_id'),
    sqlc.arg('name'),
    sqlc.arg('created_at'),
    sqlc.arg('organization_name'),
    sqlc.arg('organization_scoped')
);

-- name: FindAgentPools :many
SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)::text[]
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)::text[]
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
ORDER BY ap.created_at DESC
;

-- Find agent pools in an organization, optionally filtering by any combination of:
-- (a) name_substring: pool name contains substring
-- (b) allowed_workspace_name: workspace with name is allowed to use pool
-- (c) allowed_workspace_id: workspace with ID is allowed to use pool
--
-- name: FindAgentPoolsByOrganization :many
SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)::text[]
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)::text[]
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
LEFT JOIN (agent_pool_allowed_workspaces aw JOIN workspaces w USING (workspace_id)) ON ap.agent_pool_id = aw.agent_pool_id
WHERE ap.organization_name = sqlc.arg('organization_name')
AND   ((sqlc.arg('name_substring')::text IS NULL) OR ap.name LIKE '%' || sqlc.arg('name_substring') || '%')
AND   ((sqlc.arg('allowed_workspace_name')::text IS NULL) OR
       ap.organization_scoped OR
       w.name = sqlc.arg('allowed_workspace_name')
      )
AND   ((sqlc.arg('allowed_workspace_id')::text IS NULL) OR
       ap.organization_scoped OR
       w.workspace_id = sqlc.arg('allowed_workspace_id')
      )
GROUP BY ap.agent_pool_id
ORDER BY ap.created_at DESC
;

-- name: FindAgentPool :one
SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)::text[]
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)::text[]
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
WHERE ap.agent_pool_id = sqlc.arg('pool_id')
GROUP BY ap.agent_pool_id
;

-- name: FindAgentPoolByAgentTokenID :one
SELECT ap.*,
    (
        SELECT array_agg(w.workspace_id)::text[]
        FROM workspaces w
        WHERE w.agent_pool_id = ap.agent_pool_id
    ) AS workspace_ids,
    (
        SELECT array_agg(aw.workspace_id)::text[]
        FROM agent_pool_allowed_workspaces aw
        WHERE aw.agent_pool_id = ap.agent_pool_id
    ) AS allowed_workspace_ids
FROM agent_pools ap
JOIN agent_tokens at USING (agent_pool_id)
WHERE at.agent_token_id = sqlc.arg('agent_token_id')
GROUP BY ap.agent_pool_id
;

-- name: UpdateAgentPool :one
UPDATE agent_pools
SET name = sqlc.arg('name'),
    organization_scoped = sqlc.arg('organization_scoped')
WHERE agent_pool_id = sqlc.arg('pool_id')
RETURNING *;

-- name: DeleteAgentPool :one
DELETE
FROM agent_pools
WHERE agent_pool_id = sqlc.arg('pool_id')
RETURNING *
;

-- name: InsertAgentPoolAllowedWorkspace :exec
INSERT INTO agent_pool_allowed_workspaces (
    agent_pool_id,
    workspace_id
) VALUES (
    sqlc.arg('pool_id'),
    sqlc.arg('workspace_id')
);

-- name: DeleteAgentPoolAllowedWorkspace :exec
DELETE
FROM agent_pool_allowed_workspaces
WHERE agent_pool_id = sqlc.arg('pool_id')
AND workspace_id = sqlc.arg('workspace_id')
;

--
-- agent tokens
--

-- name: InsertAgentToken :exec
INSERT INTO agent_tokens (
    agent_token_id,
    created_at,
    description,
    agent_pool_id
) VALUES (
    sqlc.arg('agent_token_id'),
    sqlc.arg('created_at'),
    sqlc.arg('description'),
    sqlc.arg('agent_pool_id')
);

-- name: FindAgentTokenByID :one
SELECT *
FROM agent_tokens
WHERE agent_token_id = sqlc.arg('agent_token_id')
;

-- name: FindAgentTokensByAgentPoolID :many
SELECT *
FROM agent_tokens
WHERE agent_pool_id = sqlc.arg('agent_pool_id')
ORDER BY created_at DESC
;

-- name: DeleteAgentTokenByID :one
DELETE
FROM agent_tokens
WHERE agent_token_id = sqlc.arg('agent_token_id')
RETURNING agent_token_id
;
