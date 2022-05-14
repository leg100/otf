-- name: InsertStateVersion :one
INSERT INTO state_versions (
    state_version_id,
    created_at,
    updated_at,
    serial,
    state,
    run_id
) VALUES (
    pggen.arg('ID'),
    current_timestamp,
    current_timestamp,
    pggen.arg('Serial'),
    pggen.arg('State'),
    pggen.arg('RunID')
)
RETURNING created_at, updated_at
;

-- name: FindStateVersionsByWorkspaceName :many
SELECT
    state_versions.*,
    array_remove(array_agg(state_version_outputs), NULL) AS state_version_outputs,
    count(*) OVER() AS full_count
FROM state_versions
JOIN (runs JOIN workspaces USING (workspace_id)) USING (run_id)
JOIN organizations USING (organization_id)
LEFT JOIN state_version_outputs USING (state_version_id)
WHERE workspaces.name = pggen.arg('workspace_name')
AND organizations.name = pggen.arg('organization_name')
GROUP BY state_versions.state_version_id
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset')
;

-- name: FindStateVersionByID :one
SELECT
    state_versions.*,
    array_remove(array_agg(state_version_outputs), NULL) AS state_version_outputs
FROM state_versions
JOIN (runs JOIN workspaces USING (workspace_id)) USING (run_id)
LEFT JOIN state_version_outputs USING (state_version_id)
WHERE state_versions.state_version_id = pggen.arg('id')
GROUP BY state_versions.state_version_id
;

-- name: FindStateVersionLatestByWorkspaceID :one
SELECT
    state_versions.*,
    array_remove(array_agg(state_version_outputs), NULL) AS state_version_outputs
FROM state_versions
JOIN (runs JOIN workspaces USING (workspace_id)) USING (run_id)
LEFT JOIN state_version_outputs USING (state_version_id)
WHERE workspaces.workspace_id = pggen.arg('workspace_id')
GROUP BY state_versions.state_version_id
ORDER BY state_versions.serial DESC, state_versions.created_at DESC
;

-- name: DeleteStateVersionByID :exec
DELETE
FROM state_versions
WHERE state_version_id = pggen.arg('state_version_id')
;
