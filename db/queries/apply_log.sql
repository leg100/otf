-- name: InsertApplyLogChunk :one
INSERT INTO apply_logs (
    apply_id,
    chunk,
    start,
    _end,
    size
) VALUES (
    pggen.arg('ApplyID'),
    pggen.arg('Chunk'),
    pggen.arg('Start'),
    pggen.arg('End'),
    pggen.arg('Size')
)
RETURNING *;

-- name: FindApplyLogChunks :many
SELECT chunk, start, _end
FROM apply_logs
WHERE apply_id = pggen.arg('apply_id')
ORDER BY chunk_id ASC
;
