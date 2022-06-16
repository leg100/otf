-- name: InsertApply :exec
INSERT INTO applies (
    apply_id,
    job_id,
    additions,
    changes,
    destructions
) VALUES (
    pggen.arg('apply_id'),
    pggen.arg('job_id'),
    pggen.arg('additions'),
    pggen.arg('changes'),
    pggen.arg('destructions')
);

-- name: FindRunIDByApplyID :one
SELECT jobs.run_id
FROM applies
JOIN jobs USING(job_id)
WHERE applies.apply_id = pggen.arg('apply_id')
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

