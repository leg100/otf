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
    pggen.arg('notification_configuration_id'),
    pggen.arg('created_at'),
    pggen.arg('updated_at'),
    pggen.arg('name'),
    pggen.arg('url'),
    pggen.arg('triggers'),
    pggen.arg('destination_type'),
    pggen.arg('enabled'),
    pggen.arg('workspace_id')
)
;

-- name: FindNotificationConfigurations :many
SELECT *
FROM notification_configurations
WHERE workspace_id = pggen.arg('workspace_id')
;

-- name: FindNotificationConfiguration :one
SELECT *
FROM notification_configurations
WHERE notification_configuration_id = pggen.arg('notification_configuration_id')
;

-- name: FindNotificationConfigurationForUpdate :one
SELECT *
FROM notification_configurations
WHERE notification_configuration_id = pggen.arg('notification_configuration_id')
FOR UPDATE
;

-- name: UpdateNotificationConfiguration :one
UPDATE notification_configurations
SET
    updated_at = pggen.arg('updated_at'),
    enabled    = pggen.arg('enabled'),
    name       = pggen.arg('name'),
    triggers   = pggen.arg('triggers'),
    url        = pggen.arg('url')
WHERE notification_configuration_id = pggen.arg('notification_configuration_id')
RETURNING notification_configuration_id
;

-- name: DeleteNotificationConfiguration :one
DELETE FROM notification_configurations
WHERE notification_configuration_id = pggen.arg('notification_configuration_id')
RETURNING notification_configuration_id
;
