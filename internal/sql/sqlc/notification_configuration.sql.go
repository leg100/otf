// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: notification_configuration.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

const deleteNotificationConfigurationByID = `-- name: DeleteNotificationConfigurationByID :one
DELETE FROM notification_configurations
WHERE notification_configuration_id = $1
RETURNING notification_configuration_id
`

func (q *Queries) DeleteNotificationConfigurationByID(ctx context.Context, notificationConfigurationID resource.ID) (resource.ID, error) {
	row := q.db.QueryRow(ctx, deleteNotificationConfigurationByID, notificationConfigurationID)
	var notification_configuration_id resource.ID
	err := row.Scan(&notification_configuration_id)
	return notification_configuration_id, err
}

const findAllNotificationConfigurations = `-- name: FindAllNotificationConfigurations :many
SELECT notification_configuration_id, created_at, updated_at, name, url, triggers, destination_type, workspace_id, enabled
FROM notification_configurations
`

func (q *Queries) FindAllNotificationConfigurations(ctx context.Context) ([]NotificationConfiguration, error) {
	rows, err := q.db.Query(ctx, findAllNotificationConfigurations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NotificationConfiguration
	for rows.Next() {
		var i NotificationConfiguration
		if err := rows.Scan(
			&i.NotificationConfigurationID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.URL,
			&i.Triggers,
			&i.DestinationType,
			&i.WorkspaceID,
			&i.Enabled,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findNotificationConfiguration = `-- name: FindNotificationConfiguration :one
SELECT notification_configuration_id, created_at, updated_at, name, url, triggers, destination_type, workspace_id, enabled
FROM notification_configurations
WHERE notification_configuration_id = $1
`

func (q *Queries) FindNotificationConfiguration(ctx context.Context, notificationConfigurationID resource.ID) (NotificationConfiguration, error) {
	row := q.db.QueryRow(ctx, findNotificationConfiguration, notificationConfigurationID)
	var i NotificationConfiguration
	err := row.Scan(
		&i.NotificationConfigurationID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.URL,
		&i.Triggers,
		&i.DestinationType,
		&i.WorkspaceID,
		&i.Enabled,
	)
	return i, err
}

const findNotificationConfigurationForUpdate = `-- name: FindNotificationConfigurationForUpdate :one
SELECT notification_configuration_id, created_at, updated_at, name, url, triggers, destination_type, workspace_id, enabled
FROM notification_configurations
WHERE notification_configuration_id = $1
FOR UPDATE
`

func (q *Queries) FindNotificationConfigurationForUpdate(ctx context.Context, notificationConfigurationID resource.ID) (NotificationConfiguration, error) {
	row := q.db.QueryRow(ctx, findNotificationConfigurationForUpdate, notificationConfigurationID)
	var i NotificationConfiguration
	err := row.Scan(
		&i.NotificationConfigurationID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.URL,
		&i.Triggers,
		&i.DestinationType,
		&i.WorkspaceID,
		&i.Enabled,
	)
	return i, err
}

const findNotificationConfigurationsByWorkspaceID = `-- name: FindNotificationConfigurationsByWorkspaceID :many
SELECT notification_configuration_id, created_at, updated_at, name, url, triggers, destination_type, workspace_id, enabled
FROM notification_configurations
WHERE workspace_id = $1
`

func (q *Queries) FindNotificationConfigurationsByWorkspaceID(ctx context.Context, workspaceID resource.ID) ([]NotificationConfiguration, error) {
	rows, err := q.db.Query(ctx, findNotificationConfigurationsByWorkspaceID, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NotificationConfiguration
	for rows.Next() {
		var i NotificationConfiguration
		if err := rows.Scan(
			&i.NotificationConfigurationID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.URL,
			&i.Triggers,
			&i.DestinationType,
			&i.WorkspaceID,
			&i.Enabled,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertNotificationConfiguration = `-- name: InsertNotificationConfiguration :exec
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
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
)
`

type InsertNotificationConfigurationParams struct {
	NotificationConfigurationID resource.ID
	CreatedAt                   pgtype.Timestamptz
	UpdatedAt                   pgtype.Timestamptz
	Name                        pgtype.Text
	URL                         pgtype.Text
	Triggers                    []pgtype.Text
	DestinationType             pgtype.Text
	Enabled                     pgtype.Bool
	WorkspaceID                 resource.ID
}

func (q *Queries) InsertNotificationConfiguration(ctx context.Context, arg InsertNotificationConfigurationParams) error {
	_, err := q.db.Exec(ctx, insertNotificationConfiguration,
		arg.NotificationConfigurationID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Name,
		arg.URL,
		arg.Triggers,
		arg.DestinationType,
		arg.Enabled,
		arg.WorkspaceID,
	)
	return err
}

const updateNotificationConfigurationByID = `-- name: UpdateNotificationConfigurationByID :one
UPDATE notification_configurations
SET
    updated_at = $1,
    enabled    = $2,
    name       = $3,
    triggers   = $4,
    url        = $5
WHERE notification_configuration_id = $6
RETURNING notification_configuration_id
`

type UpdateNotificationConfigurationByIDParams struct {
	UpdatedAt                   pgtype.Timestamptz
	Enabled                     pgtype.Bool
	Name                        pgtype.Text
	Triggers                    []pgtype.Text
	URL                         pgtype.Text
	NotificationConfigurationID resource.ID
}

func (q *Queries) UpdateNotificationConfigurationByID(ctx context.Context, arg UpdateNotificationConfigurationByIDParams) (resource.ID, error) {
	row := q.db.QueryRow(ctx, updateNotificationConfigurationByID,
		arg.UpdatedAt,
		arg.Enabled,
		arg.Name,
		arg.Triggers,
		arg.URL,
		arg.NotificationConfigurationID,
	)
	var notification_configuration_id resource.ID
	err := row.Scan(&notification_configuration_id)
	return notification_configuration_id, err
}
