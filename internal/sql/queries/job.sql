-- name: InsertJob :exec
INSERT INTO jobs (
    run_id,
    phase,
    status
) VALUES (
    pggen.arg('run_id'),
    pggen.arg('phase'),
    pggen.arg('status')
);

-- name: FindJobs :many
SELECT
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.agent_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
;

-- name: FindJob :one
SELECT
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.agent_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE run_id = pggen.arg('run_id')
AND   phase = pggen.arg('phase')
;

-- name: FindJobForUpdate :one
SELECT
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.agent_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE run_id = pggen.arg('run_id')
AND   phase = pggen.arg('phase')
FOR UPDATE OF j
;

-- name: FindAllocatedJobs :many
SELECT
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.agent_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.agent_id = pggen.arg('agent_id')
AND   j.status = 'allocated';

-- Find signaled jobs and then immediately update signal with null.
--
-- name: FindAndUpdateSignaledJobs :many
UPDATE jobs AS j
SET signaled = NULL
FROM runs r, workspaces w
WHERE j.run_id = r.run_id
AND   r.workspace_id = w.workspace_id
AND   j.agent_id = pggen.arg('agent_id')
AND   j.status = 'running'
AND   j.signaled IS NOT NULL
RETURNING
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.agent_id,
    w.agent_pool_id,
    r.workspace_id,
    w.organization_name
;

-- name: UpdateJob :one
UPDATE jobs
SET status   = pggen.arg('status'),
    signaled = pggen.arg('signaled'),
    agent_id = pggen.arg('agent_id')
WHERE run_id = pggen.arg('run_id')
AND   phase = pggen.arg('phase')
RETURNING *;
