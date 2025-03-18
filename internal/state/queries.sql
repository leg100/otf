-- name: InsertStateVersion :exec
INSERT INTO state_versions (
    state_version_id,
    created_at,
    serial,
    state,
    status,
    workspace_id
) VALUES (
    sqlc.arg('id'),
    sqlc.arg('created_at'),
    sqlc.arg('serial'),
    sqlc.arg('state'),
    sqlc.arg('status'),
    sqlc.arg('workspace_id')
);

-- name: UpdateState :exec
UPDATE state_versions
SET state = sqlc.arg('state'), status = 'finalized'
WHERE state_version_id = sqlc.arg('state_version_id');

-- name: DiscardPendingStateVersionsByWorkspaceID :exec
UPDATE state_versions
SET status = 'discarded'
WHERE workspace_id = sqlc.arg('workspace_id')
AND status = 'pending';

-- name: FindStateVersionsByWorkspaceID :many
SELECT
    sv.*,
    (
        SELECT array_agg(svo.*)::state_version_outputs[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
        GROUP BY svo.state_version_id
    ) AS state_version_outputs
FROM state_versions sv
WHERE sv.workspace_id = sqlc.arg('workspace_id')
AND   sv.status = 'finalized'
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int
;

-- name: CountStateVersionsByWorkspaceID :one
SELECT count(*)
FROM state_versions
WHERE workspace_id = sqlc.arg('workspace_id')
AND status = 'finalized'
;

-- name: FindStateVersionByID :one
SELECT
    sv.*,
    (
        SELECT array_agg(svo.*)::state_version_outputs[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
        GROUP BY svo.state_version_id
    ) AS state_version_outputs
FROM state_versions sv
WHERE sv.state_version_id = sqlc.arg('id')
;

-- name: FindStateVersionByIDForUpdate :one
SELECT
    sv.*,
    (
        SELECT array_agg(svo.*)::state_version_outputs[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
        GROUP BY svo.state_version_id
    ) AS state_version_outputs
FROM state_versions sv
WHERE sv.state_version_id = sqlc.arg('id')
FOR UPDATE OF sv
;

-- name: FindCurrentStateVersionByWorkspaceID :one
SELECT
    sv.*,
    (
        SELECT array_agg(svo.*)::state_version_outputs[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
        GROUP BY svo.state_version_id
    ) AS state_version_outputs
FROM state_versions sv
JOIN workspaces w ON w.current_state_version_id = sv.state_version_id
WHERE w.workspace_id = sqlc.arg('workspace_id')
;

-- name: FindStateVersionStateByID :one
SELECT state
FROM state_versions
WHERE state_version_id = sqlc.arg('id')
;

-- name: DeleteStateVersionByID :one
DELETE
FROM state_versions
WHERE state_version_id = sqlc.arg('state_version_id')
RETURNING state_version_id
;

-- name: InsertStateVersionOutput :exec
INSERT INTO state_version_outputs (
    state_version_output_id,
    name,
    sensitive,
    type,
    value,
    state_version_id
) VALUES (
    sqlc.arg('id'),
    sqlc.arg('name'),
    sqlc.arg('sensitive'),
    sqlc.arg('type'),
    sqlc.arg('value'),
    sqlc.arg('state_version_id')
);

-- name: FindStateVersionOutputByID :one
SELECT *
FROM state_version_outputs
WHERE state_version_output_id = sqlc.arg('id')
;
