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

-- name: InsertLogChunk :one
INSERT INTO logs (
    run_id,
    phase,
    chunk,
    _offset
) VALUES (
    pggen.arg('run_id'),
    pggen.arg('phase'),
    pggen.arg('chunk'),
    pggen.arg('offset')
)
RETURNING chunk_id
;

-- name: FindLogChunks :one
SELECT
    substring(string_agg(chunk, '') FROM (pggen.arg('offset')+1) FOR pggen.arg('limit'))
FROM (
    SELECT run_id, phase, chunk
    FROM logs
    WHERE run_id = pggen.arg('run_id')
    AND   phase  = pggen.arg('phase')
    ORDER BY chunk_id
) c
GROUP BY run_id, phase
;

-- name: FindLogChunkByID :one
SELECT
    chunk_id,
    run_id,
    phase,
    chunk,
    _offset AS offset
FROM logs
WHERE chunk_id = pggen.arg('chunk_id')
;
