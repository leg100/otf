-- name: InsertPlanLogChunk :one
INSERT INTO logs SELECT log_id
    plan_id,
    chunk
) VALUES (
    pggen.arg('LogID'),
    pggen.arg('Chunk')
)
RETURNING *;

-- name: FindPlanLogChunks :many
SELECT chunk
FROM plan_logs
WHERE plan_id = pggen.arg('plan_id')
ORDER BY chunk_id ASC
;
