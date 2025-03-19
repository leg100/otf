-- name: InsertLogChunk :exec
INSERT INTO logs (
    chunk_id,
    run_id,
    phase,
    chunk,
    _offset
) VALUES (
    sqlc.arg('chunk_id'),
    sqlc.arg('run_id'),
    sqlc.arg('phase'),
    sqlc.arg('chunk'),
    sqlc.arg('offset')
)
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
    ORDER BY _offset
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
