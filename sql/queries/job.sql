-- name: InsertJob :exec
INSERT INTO jobs (
    job_id,
    run_id,
    status
) VALUES (
    pggen.arg('job_id'),
    pggen.arg('run_id'),
    pggen.arg('status')
);

-- name: InsertJobStatusTimestamp :exec
INSERT INTO job_status_timestamps (
    job_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    pggen.arg('Timestamp')
);

-- name: UpdateJobStatus :one
UPDATE jobs
SET status = pggen.arg('status')
WHERE job_id = pggen.arg('job_id')
RETURNING job_id
;

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
