-- name: InsertStateVersion :exec
INSERT INTO state_versions (
    state_version_id,
    created_at,
    serial,
    state,
    status,
    workspace_id
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('serial'),
    pggen.arg('state'),
    pggen.arg('status'),
    pggen.arg('workspace_id')
);

-- name: UpdateState :exec
UPDATE state_versions
SET state = pggen.arg('state'), status = 'finalized'
WHERE state_version_id = pggen.arg('state_version_id');

-- name: DiscardPendingStateVersions :exec
UPDATE state_versions
SET status = 'discarded'
WHERE status = 'pending';

-- name: FindStateVersionsByWorkspaceID :many
SELECT
    sv.*,
    array_remove(array_agg(state_version_outputs), NULL) AS state_version_outputs
FROM state_versions sv
LEFT JOIN state_version_outputs USING (state_version_id)
WHERE sv.workspace_id = pggen.arg('workspace_id')
AND   sv.status = 'finalized'
GROUP BY sv.state_version_id
ORDER BY created_at DESC
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
;

-- name: CountStateVersionsByWorkspaceID :one
SELECT count(*)
FROM state_versions
WHERE workspace_id = pggen.arg('workspace_id')
AND status = 'finalized'
;

-- name: FindStateVersionByID :one
SELECT
    state_versions.*,
    array_remove(array_agg(state_version_outputs), NULL) AS state_version_outputs
FROM state_versions
LEFT JOIN state_version_outputs USING (state_version_id)
WHERE state_versions.state_version_id = pggen.arg('id')
GROUP BY state_versions.state_version_id
;

-- name: FindCurrentStateVersionByWorkspaceID :one
SELECT
    sv.*,
    array_remove(array_agg(svo), NULL) AS state_version_outputs
FROM state_versions sv
LEFT JOIN state_version_outputs svo USING (state_version_id)
JOIN workspaces w ON w.current_state_version_id = sv.state_version_id
WHERE w.workspace_id = pggen.arg('workspace_id')
GROUP BY sv.state_version_id
;

-- name: FindStateVersionStateByID :one
SELECT state
FROM state_versions
WHERE state_version_id = pggen.arg('id')
;

-- name: DeleteStateVersionByID :one
DELETE
FROM state_versions
WHERE state_version_id = pggen.arg('state_version_id')
RETURNING state_version_id
;
