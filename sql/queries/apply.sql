-- name: InsertApply :exec
INSERT INTO applies (
    apply_id,
    run_id,
    status,
    additions,
    changes,
    destructions
) VALUES (
    pggen.arg('apply_id'),
    pggen.arg('run_id'),
    pggen.arg('status'),
    pggen.arg('additions'),
    pggen.arg('changes'),
    pggen.arg('destructions')
);

-- name: FindRunIDByApplyID :one
SELECT run_id
FROM applies
WHERE apply_id = pggen.arg('apply_id')
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

