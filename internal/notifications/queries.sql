-- name: InsertNotificationConfiguration :exec
INSERT INTO notification_configurations (
    notification_configuration_id,
    created_at,
    updated_at,
    name,
    url,
    triggers,
    destination_type,
    enabled,
    workspace_id
) VALUES (
    sqlc.arg('notification_configuration_id'),
    sqlc.arg('created_at'),
    sqlc.arg('updated_at'),
    sqlc.arg('name'),
    sqlc.arg('url'),
    sqlc.arg('triggers'),
    sqlc.arg('destination_type'),
    sqlc.arg('enabled'),
    sqlc.arg('workspace_id')
)
;

-- name: FindNotificationConfigurationsByWorkspaceID :many
SELECT *
FROM notification_configurations
WHERE workspace_id = sqlc.arg('workspace_id')
;

-- name: FindAllNotificationConfigurations :many
SELECT *
FROM notification_configurations
;

-- name: FindNotificationConfiguration :one
SELECT *
FROM notification_configurations
WHERE notification_configuration_id = sqlc.arg('notification_configuration_id')
;

-- name: FindNotificationConfigurationForUpdate :one
SELECT *
FROM notification_configurations
WHERE notification_configuration_id = sqlc.arg('notification_configuration_id')
FOR UPDATE
;

-- name: UpdateNotificationConfigurationByID :one
UPDATE notification_configurations
SET
    updated_at = sqlc.arg('updated_at'),
    enabled    = sqlc.arg('enabled'),
    name       = sqlc.arg('name'),
    triggers   = sqlc.arg('triggers'),
    url        = sqlc.arg('url')
WHERE notification_configuration_id = sqlc.arg('notification_configuration_id')
RETURNING notification_configuration_id
;

-- name: DeleteNotificationConfigurationByID :one
DELETE FROM notification_configurations
WHERE notification_configuration_id = sqlc.arg('notification_configuration_id')
RETURNING notification_configuration_id
;
