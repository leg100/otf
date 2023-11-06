-- name: AllocateJob :one
UPDATE jobs
SET agent_id = pggen.arg('agent_id')
WHERE run_id = pggen.arg('run_id')
AND   phase = pggen.arg('phase')
RETURNING *;

-- name: UpdateJobStatus :one
UPDATE jobs
SET status = pggen.arg('status')
WHERE run_id = pggen.arg('run_id')
AND   phase = pggen.arg('phase')
RETURNING *;

-- name: FindAllocatedJobs :many
SELECT
    j.run_id,
    j.phase,
    j.status,
    w.execution_mode,
    r.workspace_id,
    j.agent_id
FROM jobs j
JOIN runs r USING (run_id)
JOIN workspaces w USING (workspace_id)
WHERE j.agent_id = pggen.arg('agent_id')
AND   j.status = 'allocated'
;
