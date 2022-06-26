-- name: InsertPhaseStatusTimestamp :exec
INSERT INTO phase_status_timestamps (
    run_id,
    phase,
    status,
    timestamp
) VALUES (
    pggen.arg('run_id'),
    pggen.arg('phase'),
    pggen.arg('status'),
    pggen.arg('timestamp')
);

-- name: InsertLogChunk :exec
INSERT INTO logs (
    run_id,
    phase,
    chunk
) VALUES (
    pggen.arg('run_id'),
    pggen.arg('phase'),
    pggen.arg('chunk')
)
;

-- name: FindLogChunks :one
SELECT
    substring(string_agg(chunk, '') FROM pggen.arg('offset') FOR pggen.arg('limit'))
FROM (
    SELECT run_id, phase, chunk
    FROM logs
    WHERE run_id = pggen.arg('run_id')
    AND   phase  = pggen.arg('phase')
    ORDER BY chunk_id
) c
GROUP BY run_id, phase
;
