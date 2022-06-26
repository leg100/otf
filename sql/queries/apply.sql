-- name: InsertApply :exec
INSERT INTO applies (
    run_id,
    status
) VALUES (
    pggen.arg('run_id'),
    pggen.arg('status')
);

-- name: UpdateAppliedChangesByID :one
UPDATE applies
SET report = (
    pggen.arg('additions'),
    pggen.arg('changes'),
    pggen.arg('destructions')
)
WHERE run_id = pggen.arg('run_id')
RETURNING run_id
;

-- name: UpdateApplyStatusByID :one
UPDATE applies
SET status = pggen.arg('status')
WHERE run_id = pggen.arg('run_id')
RETURNING run_id
;

