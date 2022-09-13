-- name: InsertConfigurationVersion :exec
INSERT INTO configuration_versions (
    configuration_version_id,
    created_at,
    auto_queue_runs,
    source,
    speculative,
    status,
    workspace_id
) VALUES (
    pggen.arg('ID'),
    pggen.arg('CreatedAt'),
    pggen.arg('AutoQueueRuns'),
    pggen.arg('Source'),
    pggen.arg('Speculative'),
    pggen.arg('Status'),
    pggen.arg('WorkspaceID')
);

-- name: InsertConfigurationVersionStatusTimestamp :one
INSERT INTO configuration_version_status_timestamps (
    configuration_version_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    pggen.arg('Timestamp')
)
RETURNING *;

-- FindConfigurationVersions finds configuration_versions for a given workspace.
-- Results are paginated with limit and offset, and total count is returned.
--
-- name: FindConfigurationVersionsByWorkspaceID :many
SELECT
    configuration_versions.configuration_version_id,
    configuration_versions.created_at,
    configuration_versions.auto_queue_runs,
    configuration_versions.source,
    configuration_versions.speculative,
    configuration_versions.status,
    configuration_versions.workspace_id,
    (
        SELECT array_agg(t.*) AS configuration_version_status_timestamps
        FROM configuration_version_status_timestamps t
        WHERE t.configuration_version_id = configuration_versions.configuration_version_id
        GROUP BY configuration_version_id
    ) AS configuration_version_status_timestamps
FROM configuration_versions
JOIN workspaces USING (workspace_id)
WHERE workspaces.workspace_id = pggen.arg('workspace_id')
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset');

-- name: CountConfigurationVersionsByWorkspaceID :one
SELECT count(*)
FROM configuration_versions
WHERE configuration_versions.workspace_id = pggen.arg('workspace_id')
;

-- FindConfigurationVersionByID finds a configuration_version by its id.
--
-- name: FindConfigurationVersionByID :one
SELECT
    configuration_versions.configuration_version_id,
    configuration_versions.created_at,
    configuration_versions.auto_queue_runs,
    configuration_versions.source,
    configuration_versions.speculative,
    configuration_versions.status,
    configuration_versions.workspace_id,
    (
        SELECT array_agg(t.*) AS configuration_version_status_timestamps
        FROM configuration_version_status_timestamps t
        WHERE t.configuration_version_id = configuration_versions.configuration_version_id
        GROUP BY configuration_version_id
    ) AS configuration_version_status_timestamps
FROM configuration_versions
JOIN workspaces USING (workspace_id)
WHERE configuration_version_id = pggen.arg('configuration_version_id');

-- name: FindConfigurationVersionLatestByWorkspaceID :one
SELECT
    configuration_versions.configuration_version_id,
    configuration_versions.created_at,
    configuration_versions.auto_queue_runs,
    configuration_versions.source,
    configuration_versions.speculative,
    configuration_versions.status,
    configuration_versions.workspace_id,
    (
        SELECT array_agg(t.*) AS configuration_version_status_timestamps
        FROM configuration_version_status_timestamps t
        WHERE t.configuration_version_id = configuration_versions.configuration_version_id
        GROUP BY configuration_version_id
    ) AS configuration_version_status_timestamps
FROM configuration_versions
JOIN workspaces USING (workspace_id)
WHERE workspace_id = pggen.arg('workspace_id')
ORDER BY configuration_versions.created_at DESC;

-- name: FindConfigurationVersionByIDForUpdate :one
SELECT
    configuration_versions.configuration_version_id,
    configuration_versions.created_at,
    configuration_versions.auto_queue_runs,
    configuration_versions.source,
    configuration_versions.speculative,
    configuration_versions.status,
    configuration_versions.workspace_id,
    (
        SELECT array_agg(t.*) AS configuration_version_status_timestamps
        FROM configuration_version_status_timestamps t
        WHERE t.configuration_version_id = configuration_versions.configuration_version_id
        GROUP BY configuration_version_id
    ) AS configuration_version_status_timestamps
FROM configuration_versions
JOIN workspaces USING (workspace_id)
WHERE configuration_version_id = pggen.arg('configuration_version_id')
FOR UPDATE;

-- DownloadConfigurationVersion gets a configuration_version config
-- tarball.
--
-- name: DownloadConfigurationVersion :one
SELECT config
FROM configuration_versions
WHERE configuration_version_id = pggen.arg('configuration_version_id');

-- name: UpdateConfigurationVersionErroredByID :one
UPDATE configuration_versions
SET
    status = 'errored'
WHERE configuration_version_id = pggen.arg('id')
RETURNING configuration_version_id;

-- name: UpdateConfigurationVersionConfigByID :one
UPDATE configuration_versions
SET
    config = pggen.arg('config'),
    status = 'uploaded'
WHERE configuration_version_id = pggen.arg('id')
RETURNING configuration_version_id;

-- name: DeleteConfigurationVersionByID :one
DELETE
FROM configuration_versions
WHERE configuration_version_id = pggen.arg('id')
RETURNING configuration_version_id;
