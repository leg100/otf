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

-- FindAllocatedAndSignaledJobs finds jobs allocated to an agent that either:
-- (a) have JobAllocated status
-- (b) have JobRunning status and a non-null signal
--
-- name: FindAllocatedAndSignaledJobs :many
SELECT
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.agent_id,
    w.execution_mode,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.agent_id = pggen.arg('agent_id')
AND   j.status = 'allocated' OR (j.status = 'running' AND j.signaled IS NOT NULL)
;

-- name: FindJobs :many
SELECT
    j.run_id,
    j.phase,
    j.status,
    j.signaled,
    j.agent_id,
    w.execution_mode,
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
    w.execution_mode,
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
    w.execution_mode,
    r.workspace_id,
    w.organization_name
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE run_id = pggen.arg('run_id')
AND   phase = pggen.arg('phase')
FOR UPDATE OF j
;

-- name: UpdateJob :one
UPDATE jobs
SET status   = pggen.arg('status'),
    signaled = pggen.arg('signaled'),
    agent_id = pggen.arg('agent_id')
WHERE run_id = pggen.arg('run_id')
AND   phase = pggen.arg('phase')
RETURNING *;
