-- name: InsertJob :exec
INSERT INTO jobs (
    job_id,
    run_id
) VALUES (
    pggen.arg('job_id'),
    pggen.arg('run_id')
);

-- name: FindJobIDByApplyID :one
SELECT job_id
FROM applies
WHERE apply_id = pggen.arg('apply_id')
;

-- name: FindRunIDByJobID :one
SELECT run_id
FROM jobs
WHERE job_id = pggen.arg('job_id')
;
