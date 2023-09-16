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
	}
)

func (r row) convert() *App {
	return &App{
		ID:            r.GithubAppID.Int,
		WebhookSecret: r.WebhookSecret.String,
		PrivateKey:    r.PrivateKey.String,
	}
}

func (db *pgdb) create(ctx context.Context, app *App) error {
	_, err := db.Conn(ctx).InsertGithubApp(ctx, pggen.InsertGithubAppParams{
		GithubAppID:   pgtype.Int8{Int: app.ID, Status: pgtype.Present},
		WebhookSecret: sql.String(app.WebhookSecret),
		PrivateKey:    sql.String(app.PrivateKey),
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
	_, err := db.Conn(ctx).DeleteGithubApp(ctx)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
