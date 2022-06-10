-- name: InsertApply :exec
INSERT INTO applies (
    apply_id,
    job_id,
    status,
    additions,
    changes,
    destructions
) VALUES (
    pggen.arg('apply_id'),
    pggen.arg('job_id'),
    pggen.arg('status'),
    pggen.arg('additions'),
    pggen.arg('changes'),
    pggen.arg('destructions')
);

-- name: InsertApplyStatusTimestamp :exec
INSERT INTO apply_status_timestamps (
    apply_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    pggen.arg('Timestamp')
);

-- name: FindRunIDByApplyID :one
SELECT jobs.run_id
FROM applies
JOIN jobs USING(job_id)
WHERE applies.apply_id = pggen.arg('apply_id')
;

-- name: UpdateApplyStatus :one
UPDATE applies
SET
    status = pggen.arg('status')
WHERE apply_id = pggen.arg('id')
RETURNING apply_id
;

-- name: UpdateAppliedChangesByID :one
UPDATE applies
SET
    additions = pggen.arg('additions'),
    changes = pggen.arg('changes'),
    destructions = pggen.arg('destructions')
WHERE apply_id = pggen.arg('apply_id')
RETURNING apply_id
;

