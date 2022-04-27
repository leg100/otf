-- name: InsertPlanLogChunk :one
INSERT INTO plan_logs (
    plan_id,
    chunk,
    start,
    _end,
    size
) VALUES (
    pggen.arg('PlanID'),
    pggen.arg('Chunk'),
    pggen.arg('Start'),
    pggen.arg('End'),
    pggen.arg('Size')
)
RETURNING *;

-- name: FindPlanLogChunks :many
SELECT chunk, start, _end
FROM plan_logs
WHERE plan_id = pggen.arg('plan_id')
ORDER BY chunk_id ASC
;
