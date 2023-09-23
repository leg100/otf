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
		GithubAppID   pgtype.Int8 `json:"github_app_id"`
		WebhookSecret pgtype.Text `json:"webhook_secret"`
		PrivateKey    pgtype.Text `json:"private_key"`
		Slug          pgtype.Text `json:"slug"`
		Organization  pgtype.Text `json:"organization"`
	}
)

func (r row) convert() *App {
	app := &App{
		ID:            r.GithubAppID.Int,
		Slug:          r.Slug.String,
		WebhookSecret: r.WebhookSecret.String,
		PrivateKey:    r.PrivateKey.String,
	}
	if r.Organization.Status == pgtype.Present {
		app.Organization = &r.Organization.String
	}
	return app
}

func (db *pgdb) create(ctx context.Context, app *App) error {
	_, err := db.Conn(ctx).InsertGithubApp(ctx, pggen.InsertGithubAppParams{
		GithubAppID:   pgtype.Int8{Int: app.ID, Status: pgtype.Present},
		WebhookSecret: sql.String(app.WebhookSecret),
		PrivateKey:    sql.String(app.PrivateKey),
		Slug:          sql.String(app.Slug),
		Organization:  sql.StringPtr(app.Organization),
	})
	return err
}

func (db *pgdb) get(ctx context.Context) (*App, error) {
	result, err := db.Conn(ctx).FindGithubApp(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	return row(result).convert(), nil
}

func (db *pgdb) delete(ctx context.Context) error {
	return db.Lock(ctx, "github_apps", func(ctx context.Context, q pggen.Querier) error {
		result, err := db.Conn(ctx).FindGithubApp(ctx)
		if err != nil {
			return sql.Error(err)
		}
		_, err = db.Conn(ctx).DeleteGithubApp(ctx, result.GithubAppID)
		if err != nil {
			return sql.Error(err)
		}
		return nil
	})
}
