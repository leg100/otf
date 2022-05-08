-- name: InsertPlanLogChunk :one
INSERT INTO plan_logs (
    plan_id,
    chunk
) VALUES (
    pggen.arg('PlanID'),
    pggen.arg('Chunk')
)
RETURNING *;

-- name: FindPlanLogChunks :one
SELECT string_agg(chunk, '')
FROM (
    SELECT plan_id, chunk
    FROM plan_logs
    WHERE plan_id = pggen.arg('plan_id')
    ORDER BY chunk_id
) c
GROUP BY plan_id
;
