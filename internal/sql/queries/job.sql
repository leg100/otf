-- name: ReallocateJob :one
UPDATE jobs
SET agent_id = pggen.arg('agent_id')
WHERE run_id = pggen.arg('run_id')
AND   phase = pggen.arg('phase');
