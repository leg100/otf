package notifications

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a notification configuration database on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) create(ctx context.Context, nc *Config) error {
	_, err := db.Exec(ctx, `
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
    @id,
    @created_at,
    @updated_at,
    @name,
    @url,
    @triggers,
    @destination_type,
    @enabled,
    @workspace_id
)
`,
		pgx.NamedArgs{
			"id":               nc.ID,
			"created_at":       nc.CreatedAt,
			"updated_at":       nc.UpdatedAt,
			"name":             nc.Name,
			"enabled":          nc.Enabled,
			"destination_type": nc.DestinationType,
			"workspace_id":     nc.WorkspaceID,
			"triggers":         nc.Triggers,
			"url":              nc.URL,
		},
	)
	return err
}

func (db *pgdb) update(ctx context.Context, id resource.TfeID, updateFunc func(context.Context, *Config) error) (*Config, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Config, error) {
			rows := db.Query(ctx, `
SELECT notification_configuration_id, created_at, updated_at, name, url, triggers, destination_type, workspace_id, enabled
FROM notification_configurations
WHERE notification_configuration_id = $1
FOR UPDATE
`, id)
			return sql.CollectOneRow(rows, db.scan)
		},
		updateFunc,
		func(ctx context.Context, conn sql.Connection, nc *Config) error {
			_, err := db.Exec(ctx, `
UPDATE notification_configurations
SET
    updated_at = @updated_at,
    enabled    = @enabled,
    name       = @name,
    triggers   = @triggers,
    url        = @url
WHERE notification_configuration_id = @id
RETURNING notification_configuration_id
`,
				pgx.NamedArgs{
					"id":         nc.ID,
					"updated_at": nc.UpdatedAt,
					"name":       nc.Name,
					"enabled":    nc.Enabled,
					"triggers":   nc.Triggers,
					"url":        nc.URL,
				},
			)
			return err
		},
	)
}

func (db *pgdb) list(ctx context.Context, workspaceID resource.TfeID) ([]*Config, error) {
	rows := db.Query(ctx, `
SELECT notification_configuration_id, created_at, updated_at, name, url, triggers, destination_type, workspace_id, enabled
FROM notification_configurations
WHERE workspace_id = $1
`, workspaceID)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) listAll(ctx context.Context) ([]*Config, error) {
	rows := db.Query(ctx, `
SELECT notification_configuration_id, created_at, updated_at, name, url, triggers, destination_type, workspace_id, enabled
FROM notification_configurations
`)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) get(ctx context.Context, id resource.TfeID) (*Config, error) {
	rows := db.Query(ctx, `
SELECT notification_configuration_id, created_at, updated_at, name, url, triggers, destination_type, workspace_id, enabled
FROM notification_configurations
WHERE notification_configuration_id = $1
`, id)
	return sql.CollectOneRow(rows, db.scan)
}

func (db *pgdb) delete(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE FROM notification_configurations
WHERE notification_configuration_id = $1
RETURNING notification_configuration_id
`, id)
	return err
}

func (db *pgdb) scan(row pgx.CollectableRow) (*Config, error) {
	return pgx.RowToAddrOfStructByName[Config](row)
}
