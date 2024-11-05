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
WHERE j.run_id = sqlc.arg('run_id')
AND   phase = sqlc.arg('phase')
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
WHERE run_id = sqlc.arg('run_id')
AND   phase = sqlc.arg('phase')
RETURNING *;
