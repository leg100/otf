-- name: InsertJob :exec
INSERT INTO jobs (
    job_id,
    created_at,
    status
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('status')
);
