package notifications

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type (
	// pgdb is a notification configuration database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	pgresult struct {
		NotificationConfigurationID pgtype.Text        `json:"notification_configuration_id"`
		CreatedAt                   pgtype.Timestamptz `json:"created_at"`
		UpdatedAt                   pgtype.Timestamptz `json:"updated_at"`
		Name                        pgtype.Text        `json:"name"`
		URL                         pgtype.Text        `json:"url"`
		Triggers                    []string           `json:"triggers"`
		DestinationType             pgtype.Text        `json:"destination_type"`
		WorkspaceID                 pgtype.Text        `json:"workspace_id"`
		Enabled                     pgtype.Bool        `json:"enabled"`
	}
)

func (r pgresult) toNotificationConfiguration() *Config {
	nc := &Config{
		ID:              r.NotificationConfigurationID.String,
		CreatedAt:       r.CreatedAt.Time.UTC(),
		UpdatedAt:       r.UpdatedAt.Time.UTC(),
		Name:            r.Name.String,
		Enabled:         r.Enabled.Bool,
		DestinationType: Destination(r.DestinationType.String),
		WorkspaceID:     r.WorkspaceID.String,
	}
	for _, t := range r.Triggers {
		nc.Triggers = append(nc.Triggers, Trigger(t))
	}
	if r.URL.Status == pgtype.Present {
		nc.URL = &r.URL.String
	}
	return nc
}

func (db *pgdb) create(ctx context.Context, nc *Config) error {
	params := sqlc.InsertNotificationConfigurationParams{
		NotificationConfigurationID: sql.String(nc.ID),
		CreatedAt:                   sql.Timestamptz(nc.CreatedAt),
		UpdatedAt:                   sql.Timestamptz(nc.UpdatedAt),
		Name:                        sql.String(nc.Name),
		Enabled:                     sql.Bool(nc.Enabled),
		DestinationType:             sql.String(string(nc.DestinationType)),
		URL:                         sql.NullString(),
		WorkspaceID:                 sql.String(nc.WorkspaceID),
	}
	for _, t := range nc.Triggers {
		params.Triggers = append(params.Triggers, string(t))
	}
	if nc.URL != nil {
		params.URL = sql.String(*nc.URL)
	}
	_, err := db.Conn(ctx).InsertNotificationConfiguration(ctx, params)
	return sql.Error(err)
}

func (db *pgdb) update(ctx context.Context, id string, updateFunc func(*Config) error) (*Config, error) {
	var nc *Config
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		result, err := q.FindNotificationConfigurationForUpdate(ctx, sql.String(id))
		if err != nil {
			return sql.Error(err)
		}
		nc = pgresult(result).toNotificationConfiguration()
		if err := updateFunc(nc); err != nil {
			return sql.Error(err)
		}
		params := sqlc.UpdateNotificationConfigurationByIDParams{
			UpdatedAt:                   sql.Timestamptz(internal.CurrentTimestamp(nil)),
			Enabled:                     sql.Bool(nc.Enabled),
			Name:                        sql.String(nc.Name),
			URL:                         sql.NullString(),
			NotificationConfigurationID: sql.String(nc.ID),
		}
		for _, t := range nc.Triggers {
			params.Triggers = append(params.Triggers, string(t))
		}
		if nc.URL != nil {
			params.URL = sql.String(*nc.URL)
		}
		_, err = q.UpdateNotificationConfigurationByID(ctx, params)
		return err
	})
	return nc, err
}

func (db *pgdb) list(ctx context.Context, workspaceID string) ([]*Config, error) {
	results, err := db.Conn(ctx).FindNotificationConfigurationsByWorkspaceID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	configs := make([]*Config, len(results))
	for i, row := range results {
		configs[i] = pgresult(row).toNotificationConfiguration()
	}
	return configs, nil
}

func (db *pgdb) listAll(ctx context.Context) ([]*Config, error) {
	results, err := db.Conn(ctx).FindAllNotificationConfigurations(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}

	configs := make([]*Config, len(results))
	for i, row := range results {
		configs[i] = pgresult(row).toNotificationConfiguration()
	}
	return configs, nil
}

func (db *pgdb) get(ctx context.Context, id string) (*Config, error) {
	row, err := db.Conn(ctx).FindNotificationConfiguration(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(row).toNotificationConfiguration(), nil
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteNotificationConfigurationByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
