package notifications

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is a notification configuration database on postgres
	pgdb struct {
		internal.DB // provides access to generated SQL queries
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
		Enabled                     bool               `json:"enabled"`
	}
)

func (r pgresult) toNotificationConfiguration() *Config {
	nc := &Config{
		ID:              r.NotificationConfigurationID.String,
		CreatedAt:       r.CreatedAt.Time.UTC(),
		UpdatedAt:       r.UpdatedAt.Time.UTC(),
		Name:            r.Name.String,
		URL:             r.URL.String,
		DestinationType: Destination(r.DestinationType.Status),
		WorkspaceID:     r.WorkspaceID.String,
	}
	for _, t := range r.Triggers {
		nc.Triggers = append(nc.Triggers, Trigger(t))
	}
	return nc
}

func (db *pgdb) create(ctx context.Context, nc *Config) error {
	params := pggen.InsertNotificationConfigurationParams{
		NotificationConfigurationID: sql.String(nc.ID),
		CreatedAt:                   sql.Timestamptz(nc.CreatedAt),
		UpdatedAt:                   sql.Timestamptz(nc.UpdatedAt),
		Name:                        sql.String(nc.Name),
		WorkspaceID:                 sql.String(nc.WorkspaceID),
	}
	for _, t := range nc.Triggers {
		params.Triggers = append(params.Triggers, string(t))
	}
	_, err := db.InsertNotificationConfiguration(ctx, params)
	return sql.Error(err)
}

func (db *pgdb) update(ctx context.Context, id string, updateFunc func(*Config) error) (*Config, error) {
	var nc *Config
	err := db.Tx(ctx, func(tx internal.DB) error {
		result, err := tx.FindNotificationConfigurationForUpdate(ctx, sql.String(id))
		if err != nil {
			return sql.Error(err)
		}
		nc = pgresult(result).toNotificationConfiguration()
		if err := updateFunc(nc); err != nil {
			return sql.Error(err)
		}
		params := pggen.UpdateNotificationConfigurationParams{
			Name:    sql.String(nc.Name),
			Enabled: nc.Enabled,
			URL:     sql.String(nc.URL),
		}
		for _, t := range nc.Triggers {
			params.Triggers = append(params.Triggers, string(t))
		}
		_, err = tx.UpdateNotificationConfiguration(ctx, params)
		return err
	})
	return nc, err
}

func (db *pgdb) list(ctx context.Context, workspaceID string) ([]*Config, error) {
	results, err := db.FindNotificationConfigurations(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	var configs []*Config
	for _, row := range results {
		configs = append(configs, pgresult(row).toNotificationConfiguration())
	}
	return configs, nil
}

func (db *pgdb) get(ctx context.Context, id string) (*Config, error) {
	row, err := db.FindNotificationConfiguration(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(row).toNotificationConfiguration(), nil
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	_, err := db.DeleteNotificationConfiguration(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
