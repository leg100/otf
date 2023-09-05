package github

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is a github app database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// row represents a database row for a github app
	row struct {
		GithubAppID      pgtype.Text `json:"github_app_id"`
		WebhookSecret    pgtype.Text `json:"webhook_secret"`
		Pem              pgtype.Text `json:"pem"`
		OrganizationName pgtype.Text `json:"organization_name"`
	}
)

func (db *pgdb) create(ctx context.Context, app *ghapp) error {
	_, err := db.Conn(ctx).InsertGithubApp(ctx, pggen.InsertGithubAppParams{
		GithubAppID:      sql.String(app.ID),
		WebhookSecret:    sql.String(app.WebhookSecret),
		Pem:              sql.String(app.Pem),
		OrganizationName: sql.String(app.Organization),
	})
	return err
}

func (db *pgdb) get(ctx context.Context, id string) (*ghapp, error) {
	row, err := db.Conn(ctx).FindVCSProvider(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return &ghapp{
		ID: row.GithubAppID.String,
	}, nil
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteGithubAppByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
