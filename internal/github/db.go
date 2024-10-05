package github

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type (
	// pgdb is a github app database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// AppRow represents a database AppRow for a github app
	AppRow struct {
		GithubAppID   pgtype.Int8 `json:"github_app_id"`
		WebhookSecret pgtype.Text `json:"webhook_secret"`
		PrivateKey    pgtype.Text `json:"private_key"`
		Slug          pgtype.Text `json:"slug"`
		Organization  pgtype.Text `json:"organization"`
	}

	// AppInstallRow is a database row for a github app install
	AppInstallRow struct {
		GithubAppID   pgtype.Int8 `json:"github_app_id"`
		InstallID     pgtype.Int8 `json:"install_id"`
		Username      pgtype.Text `json:"username"`
		Organization  pgtype.Text `json:"organization"`
		VCSProviderID pgtype.Text `json:"vcs_provider_id"`
	}
)

func (r AppRow) convert() *App {
	app := &App{
		ID:            r.GithubAppID.Int64,
		Slug:          r.Slug.String,
		WebhookSecret: r.WebhookSecret.String,
		PrivateKey:    r.PrivateKey.String,
	}
	if r.Organization.Valid {
		app.Organization = &r.Organization.String
	}
	return app
}

func (db *pgdb) create(ctx context.Context, app *App) error {
	err := db.Querier(ctx).InsertGithubApp(ctx, sqlc.InsertGithubAppParams{
		GithubAppID:   pgtype.Int8{Int64: app.ID, Valid: true},
		WebhookSecret: sql.String(app.WebhookSecret),
		PrivateKey:    sql.String(app.PrivateKey),
		Slug:          sql.String(app.Slug),
		Organization:  sql.StringPtr(app.Organization),
	})
	return err
}

func (db *pgdb) get(ctx context.Context) (*App, error) {
	result, err := db.Querier(ctx).FindGithubApp(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	return AppRow(result).convert(), nil
}

func (db *pgdb) delete(ctx context.Context) error {
	return db.Lock(ctx, "github_apps", func(ctx context.Context, q *sqlc.Queries) error {
		result, err := db.Querier(ctx).FindGithubApp(ctx)
		if err != nil {
			return sql.Error(err)
		}
		_, err = db.Querier(ctx).DeleteGithubApp(ctx, result.GithubAppID)
		if err != nil {
			return sql.Error(err)
		}
		return nil
	})
}
