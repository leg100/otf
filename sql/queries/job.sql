-- name: FindQueuedJobs :many
SELECT
    j.job_id,
    j.run_id,
    p.relname AS job_type,
    j.status,
    r.is_destroy,
    r.refresh,
    r.refresh_only,
    w.auto_apply,
    cv.speculative,
    r.configuration_version_id,
    r.workspace_id
FROM jobs j
JOIN pg_class p ON j.tableoid = p.tableoid
JOIN runs r ON r.run_id = j.run_id
JOIN configuration_versions cv USING(configuration_version_id)
JOIN workspaces w ON r.workspace_id = w.workspace_id
WHERE j.status = 'queued'
AND   j.tableoid = p.oid
;

-- name: InsertJobStatusTimestamp :exec
INSERT INTO job_status_timestamps (
    job_id,
    status,
    timestamp
) VALUES (
    pggen.arg('job_id'),
    pggen.arg('status'),
    pggen.arg('timestamp')
);

-- name: InsertLogChunk :exec
INSERT INTO logs (
    job_id,
    chunk
) VALUES (
    pggen.arg('job_id'),
    pggen.arg('chunk')
)
;

-- name: FindLogChunks :one
SELECT
    substring(string_agg(chunk, '') FROM pggen.arg('offset') FOR pggen.arg('limit'))
FROM (
    SELECT job_id, chunk
    FROM logs
    WHERE job_id = pggen.arg('job_id')
    ORDER BY chunk_id
) c
GROUP BY job_id
;

-- name: UpdateJobStatus :one
UPDATE jobs
SET status = pggen.arg('status')
WHERE job_id = pggen.arg('job_id')
RETURNING job_id
;

