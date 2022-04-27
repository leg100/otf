-- InsertConfigurationVersion inserts a configuration version and
-- returns the entire row.
--
-- name: InsertConfigurationVersion :one
INSERT INTO configuration_versions (
    configuration_version_id,
    created_at,
    updated_at,
    auto_queue_runs,
    source,
    speculative,
    status
) VALUES (
    pggen.arg('ID'),
    NOW(),
    NOW(),
    pggen.arg('AutoQueueRuns'),
    pggen.arg('Source'),
    pggen.arg('Speculative'),
    pggen.arg('Status')
)
RETURNING *;

-- name: InsertConfigurationVersionStatusTimestamp :one
INSERT INTO configuration_version_status_timestamps (
    configuration_version_id,
    status,
    timestamp
) VALUES (
    pggen.arg('ID'),
    pggen.arg('Status'),
    NOW()
)
RETURNING *;

-- FindConfigurationVersions finds configuration_versions for a given workspace.
-- Results are paginated with limit and offset, and total count is returned.
--
-- name: FindConfigurationVersions :many
SELECT configuration_versions.*, (workspaces.*)::"workspaces" AS workspace, count(*) OVER() AS full_count
FROM configuration_versions
JOIN workspaces USING (workspace_id)
WHERE workspaces.workspace_id = pggen.arg('workspace_id')
LIMIT pggen.arg('limit')
OFFSET pggen.arg('offset');

-- FindConfigurationVersionByID finds a configuration_version by its id.
--
-- name: FindConfigurationVersionByID :one
SELECT
    configuration_versions.configuration_version_id,
    configuration_versions.created_at,
    configuration_versions.updated_at,
    configuration_versions.auto_queue_runs,
    configuration_versions.source,
    configuration_versions.speculative,
    configuration_versions.status,
    (workspaces.*)::"workspaces" AS workspace
FROM configuration_versions
JOIN workspaces USING (workspace_id)
WHERE configuration_version_id = pggen.arg('configuration_version_id');

-- DownloadConfigurationVersion gets a configuration_version config
-- tarball.
--
-- name: DownloadConfigurationVersion :one
SELECT config
FROM configuration_versions
WHERE configuration_version_id = pggen.arg('configuration_version_id');

-- UploadConfigurationVersion sets a config tarball on a configuration version,
-- and sets the status to uploaded.
--
-- name: UploadConfigurationVersion :one
UPDATE configuration_versions
SET
    config = pggen.arg('config'),
    status = 'uploaded',
    updated_at = NOW()
WHERE configuration_version_id = pggen.arg('id')
RETURNING *;
