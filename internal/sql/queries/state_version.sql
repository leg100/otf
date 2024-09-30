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
        SELECT array_agg(svo.state_version_output_id)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_ids,
    (
        SELECT array_agg(svo.sensitive)::bool[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_sensitives,
    (
        SELECT array_agg(svo.type)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_types,
    (
        SELECT array_agg(svo.value)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_values,
    (
        SELECT array_agg(svo.name)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_names
FROM state_versions sv
WHERE sv.workspace_id = sqlc.arg('workspace_id')
AND   sv.status = 'finalized'
GROUP BY sv.state_version_id
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset')
;

-- name: CountStateVersionsByWorkspaceID :one
SELECT count(*)
FROM state_versions
WHERE workspace_id = sqlc.arg('workspace_id')
AND status = 'finalized'
;

-- name: FindStateVersionByID :one
SELECT
    state_versions.*,
    (
        SELECT array_agg(svo.state_version_output_id)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_ids,
    (
        SELECT array_agg(svo.sensitive)::bool[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_sensitives,
    (
        SELECT array_agg(svo.type)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_types,
    (
        SELECT array_agg(svo.value)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_values,
    (
        SELECT array_agg(svo.name)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_names
FROM state_versions
WHERE state_versions.state_version_id = sqlc.arg('id')
GROUP BY state_versions.state_version_id
;

-- name: FindStateVersionByIDForUpdate :one
SELECT
    sv.*,
    (
        SELECT array_agg(svo.state_version_output_id)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_ids,
    (
        SELECT array_agg(svo.sensitive)::bool[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_sensitives,
    (
        SELECT array_agg(svo.type)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_types,
    (
        SELECT array_agg(svo.value)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_values,
    (
        SELECT array_agg(svo.name)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_names
FROM state_versions sv
WHERE sv.state_version_id = sqlc.arg('id')
FOR UPDATE OF sv
;

-- name: FindCurrentStateVersionByWorkspaceID :one
SELECT
    sv.*,
    (
        SELECT array_agg(svo.state_version_output_id)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_ids,
    (
        SELECT array_agg(svo.sensitive)::bool[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_sensitives,
    (
        SELECT array_agg(svo.type)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_types,
    (
        SELECT array_agg(svo.value)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_values,
    (
        SELECT array_agg(svo.name)::text[]
        FROM state_version_outputs svo
        WHERE svo.state_version_id = sv.state_version_id
    ) AS state_version_output_names
FROM state_versions sv
JOIN workspaces w ON w.current_state_version_id = sv.state_version_id
WHERE w.workspace_id = sqlc.arg('workspace_id')
GROUP BY sv.state_version_id
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
