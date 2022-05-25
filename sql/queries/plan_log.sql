-- name: InsertPlanLogChunk :one
INSERT INTO plan_logs (
    plan_id,
    chunk
) VALUES (
    pggen.arg('plan_id'),
    pggen.arg('chunk')
)
RETURNING *;

-- name: FindPlanLogChunks :one
SELECT
    substring(string_agg(chunk, '') FROM pggen.arg('offset') FOR pggen.arg('limit'))
FROM (
    SELECT plan_id, chunk
    FROM plan_logs
    WHERE plan_id = pggen.arg('plan_id')
    ORDER BY chunk_id
) c
GROUP BY plan_id
;
