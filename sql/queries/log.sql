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

-- name: FindAllLogChunksUsingApplyID :one
SELECT
    string_agg(chunk, '')
FROM (
    SELECT logs.job_id, logs.chunk
    FROM logs
    JOIN applies USING(job_id)
    WHERE applies.apply_id = pggen.arg('apply_id')
    ORDER BY chunk_id
) c
GROUP BY job_id
;
