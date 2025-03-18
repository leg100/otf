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
    sqlc.arg('id'),
    sqlc.arg('created_at'),
    sqlc.arg('auto_queue_runs'),
    sqlc.arg('source'),
    sqlc.arg('speculative'),
    sqlc.arg('status'),
    sqlc.arg('workspace_id')
);

-- name: InsertConfigurationVersionStatusTimestamp :one
INSERT INTO configuration_version_status_timestamps (
    configuration_version_id,
    status,
    timestamp
) VALUES (
    sqlc.arg('id'),
    sqlc.arg('status'),
    sqlc.arg('timestamp')
)
RETURNING *;

-- FindConfigurationVersions finds configuration_versions for a given workspace.
-- Results are paginated with limit and offset, and total count is returned.
--
-- name: FindConfigurationVersionsByWorkspaceID :many
SELECT
    cv.configuration_version_id,
    cv.created_at,
    cv.auto_queue_runs,
    cv.source,
    cv.speculative,
    cv.status,
    cv.workspace_id,
    (
        SELECT array_agg(cst.*)::configuration_version_status_timestamps[]
        FROM configuration_version_status_timestamps cst
        WHERE cst.configuration_version_id = cv.configuration_version_id
        GROUP BY cst.configuration_version_id
    ) AS status_timestamps,
    ia::"ingress_attributes" AS ingress_attributes
FROM configuration_versions cv
JOIN workspaces USING (workspace_id)
LEFT JOIN ingress_attributes ia USING (configuration_version_id)
WHERE workspaces.workspace_id = sqlc.arg('workspace_id')
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int
;

-- name: CountConfigurationVersionsByWorkspaceID :one
SELECT count(*)
FROM configuration_versions
WHERE configuration_versions.workspace_id = sqlc.arg('workspace_id')
;

-- FindConfigurationVersionByID finds a configuration_version by its id.
--
-- name: FindConfigurationVersionByID :one
SELECT
    cv.configuration_version_id,
    cv.created_at,
    cv.auto_queue_runs,
    cv.source,
    cv.speculative,
    cv.status,
    cv.workspace_id,
    (
        SELECT array_agg(cst.*)::configuration_version_status_timestamps[]
        FROM configuration_version_status_timestamps cst
        WHERE cst.configuration_version_id = cv.configuration_version_id
        GROUP BY cst.configuration_version_id
    ) AS status_timestamps,
    ia::ingress_attributes AS ingress_attributes
FROM configuration_versions cv
JOIN workspaces USING (workspace_id)
LEFT JOIN ingress_attributes ia USING(configuration_version_id)
WHERE cv.configuration_version_id = sqlc.arg('configuration_version_id')
;

-- name: FindConfigurationVersionLatestByWorkspaceID :one
SELECT
    cv.configuration_version_id,
    cv.created_at,
    cv.auto_queue_runs,
    cv.source,
    cv.speculative,
    cv.status,
    cv.workspace_id,
    (
        SELECT array_agg(cst.*)::configuration_version_status_timestamps[]
        FROM configuration_version_status_timestamps cst
        WHERE cst.configuration_version_id = cv.configuration_version_id
        GROUP BY cst.configuration_version_id
    ) AS status_timestamps,
    ia::ingress_attributes AS ingress_attributes
FROM configuration_versions cv
JOIN workspaces USING (workspace_id)
LEFT JOIN ingress_attributes ia USING(configuration_version_id)
WHERE cv.workspace_id = sqlc.arg('workspace_id')
ORDER BY cv.created_at DESC
;

-- name: FindConfigurationVersionByIDForUpdate :one
SELECT
    cv.configuration_version_id,
    cv.created_at,
    cv.auto_queue_runs,
    cv.source,
    cv.speculative,
    cv.status,
    cv.workspace_id,
    (
        SELECT array_agg(cst.*)::configuration_version_status_timestamps[]
        FROM configuration_version_status_timestamps cst
        WHERE cst.configuration_version_id = cv.configuration_version_id
        GROUP BY cst.configuration_version_id
    ) AS status_timestamps,
    ia::ingress_attributes AS ingress_attributes
FROM configuration_versions cv
JOIN workspaces USING (workspace_id)
LEFT JOIN ingress_attributes ia USING(configuration_version_id)
WHERE cv.configuration_version_id = sqlc.arg('configuration_version_id')
FOR UPDATE OF cv;

-- DownloadConfigurationVersion gets a configuration_version config
-- tarball.
--
-- name: DownloadConfigurationVersion :one
SELECT config
FROM configuration_versions
WHERE configuration_version_id = sqlc.arg('configuration_version_id')
AND   status                   = 'uploaded';

-- name: UpdateConfigurationVersionErroredByID :one
UPDATE configuration_versions
SET
    status = 'errored'
WHERE configuration_version_id = sqlc.arg('id')
RETURNING configuration_version_id;

-- name: UpdateConfigurationVersionConfigByID :one
UPDATE configuration_versions
SET
    config = sqlc.arg('config'),
    status = 'uploaded'
WHERE configuration_version_id = sqlc.arg('id')
RETURNING configuration_version_id;

-- name: DeleteConfigurationVersionByID :one
DELETE
FROM configuration_versions
WHERE configuration_version_id = sqlc.arg('id')
RETURNING configuration_version_id;

-- name: InsertIngressAttributes :exec
INSERT INTO ingress_attributes (
    branch,
    commit_sha,
    commit_url,
    pull_request_number,
    pull_request_url,
    pull_request_title,
    sender_username,
    sender_avatar_url,
    sender_html_url,
    identifier,
    tag,
    is_pull_request,
    on_default_branch,
    configuration_version_id
) VALUES (
    sqlc.arg('branch'),
    sqlc.arg('commit_sha'),
    sqlc.arg('commit_url'),
    sqlc.arg('pull_request_number'),
    sqlc.arg('pull_request_url'),
    sqlc.arg('pull_request_title'),
    sqlc.arg('sender_username'),
    sqlc.arg('sender_avatar_url'),
    sqlc.arg('sender_html_url'),
    sqlc.arg('identifier'),
    sqlc.arg('tag'),
    sqlc.arg('is_pull_request'),
    sqlc.arg('on_default_branch'),
    sqlc.arg('configuration_version_id')
);
