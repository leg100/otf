-- name: InsertApply :exec
INSERT INTO applies (
    run_id,
    status
) VALUES (
    $1,
    $2
);

-- name: UpdateAppliedChangesByID :one
UPDATE applies
SET resource_report = (
    sqlc.arg(additions)::int,
    sqlc.arg(changes)::int,
    sqlc.arg(destructions)::int
)
WHERE run_id = $1
RETURNING run_id
;

-- name: UpdateApplyStatusByID :one
UPDATE applies
SET status = $2
WHERE run_id = $1
RETURNING run_id
;
