-- name: InsertPhaseStatusTimestamp :exec
INSERT INTO phase_status_timestamps (
    run_id,
    phase,
    status,
    timestamp
) VALUES (
    sqlc.arg('run_id'),
    sqlc.arg('phase'),
    sqlc.arg('status'),
    sqlc.arg('timestamp')
);

-- name: InsertLogChunk :one
INSERT INTO logs (
    run_id,
    phase,
    chunk,
    _offset
) VALUES (
    sqlc.arg('run_id'),
    sqlc.arg('phase'),
    sqlc.arg('chunk'),
    sqlc.arg('offset')
)
RETURNING chunk_id
;

-- FindLogs retrieves all the logs for the given run and phase.
--
-- name: FindLogs :one
SELECT
    string_agg(chunk, '')
FROM (
    SELECT run_id, phase, chunk
    FROM logs
    WHERE run_id = sqlc.arg('run_id')
    AND   phase  = sqlc.arg('phase')
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
WHERE chunk_id = sqlc.arg('chunk_id')
;
