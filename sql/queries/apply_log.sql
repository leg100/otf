-- name: InsertApplyLogChunk :exec
INSERT INTO apply_logs (
    apply_id,
    chunk
) VALUES (
    pggen.arg('apply_id'),
    pggen.arg('chunk')
)
;

-- name: FindApplyLogChunks :one
SELECT
    substring(string_agg(chunk, '') FROM pggen.arg('offset') FOR pggen.arg('limit'))
FROM (
    SELECT apply_id, chunk
    FROM apply_logs
    WHERE apply_id = pggen.arg('apply_id')
    ORDER BY chunk_id
) c
GROUP BY apply_id
;
