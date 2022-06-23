-- name: InsertApply :exec
INSERT INTO applies (
    apply_id,
    run_id,
    status
) VALUES (
    pggen.arg('apply_id'),
    pggen.arg('run_id'),
    pggen.arg('status')
);

-- name: InsertApplyStatusTimestamp :exec
INSERT INTO apply_status_timestamps (
    apply_id,
    status,
    timestamp
) VALUES (
    pggen.arg('apply_id'),
    pggen.arg('status'),
    pggen.arg('timestamp')
);

-- name: FindRunIDByApplyID :one
SELECT run_id
FROM applies
WHERE apply_id = pggen.arg('apply_id')
;

-- name: UpdateAppliedChangesByID :one
UPDATE applies
SET report = (
    pggen.arg('additions'),
    pggen.arg('changes'),
    pggen.arg('destructions')
)
WHERE apply_id = pggen.arg('apply_id')
RETURNING apply_id
;

-- name: UpdateApplyStatusByID :one
UPDATE applies
SET status = pggen.arg('status')
WHERE apply_id = pggen.arg('apply_id')
RETURNING apply_id
;

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
