-- name: InsertStateVersion :exec
INSERT INTO state_versions (
    state_version_id,
    created_at,
    serial,
    state,
    workspace_id
) VALUES (
    pggen.arg('id'),
    pggen.arg('created_at'),
    pggen.arg('serial'),
    pggen.arg('state'),
    pggen.arg('workspace_id')
);

-- name: FindStateVersionsByWorkspaceName :many
SELECT
    state_versions.*,
    array_remove(array_agg(state_version_outputs), NULL) AS state_version_outputs
FROM state_versions
JOIN workspaces USING (workspace_id)
LEFT JOIN state_version_outputs USING (state_version_id)
WHERE workspaces.name               = pggen.arg('workspace_name')
AND   workspaces.organization_name  = pggen.arg('organization_name')
GROUP BY state_versions.state_version_id
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
;

-- name: CountStateVersionsByWorkspaceName :one
SELECT count(*)
FROM state_versions
JOIN workspaces USING (workspace_id)
WHERE workspaces.name                 = pggen.arg('workspace_name')
AND   workspaces.organization_name    = pggen.arg('organization_name')
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
