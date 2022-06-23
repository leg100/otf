-- name: InsertApply :exec
INSERT INTO applies (
    apply_id,
    job_id
) VALUES (
    pggen.arg('apply_id'),
    pggen.arg('job_id')
);

-- name: FindRunIDByApplyID :one
SELECT jobs.run_id
FROM applies
JOIN jobs USING(job_id)
WHERE applies.apply_id = pggen.arg('apply_id')
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

