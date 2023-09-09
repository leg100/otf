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
		GithubAppID   pgtype.Text `json:"github_app_id"`
		AppID         pgtype.Int8 `json:"app_id"`
		WebhookSecret pgtype.Text `json:"webhook_secret"`
		PrivateKey    pgtype.Text `json:"private_key"`
	}
)

func (r row) convert() *GithubApp {
	return &GithubApp{
		ID:            r.GithubAppID.String,
		AppID:         r.AppID.Int,
		WebhookSecret: r.WebhookSecret.String,
		PrivateKey:    r.PrivateKey.String,
	}
}

func (db *pgdb) create(ctx context.Context, app *GithubApp) error {
	_, err := db.Conn(ctx).InsertGithubApp(ctx, pggen.InsertGithubAppParams{
		GithubAppID:   sql.String(app.ID),
		AppID:         pgtype.Int8{Int: app.AppID, Status: pgtype.Present},
		WebhookSecret: sql.String(app.WebhookSecret),
		PrivateKey:    sql.String(app.PrivateKey),
	})
	return err
}

func (db *pgdb) get(ctx context.Context) (*GithubApp, error) {
	result, err := db.Conn(ctx).FindGithubApps(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	if len(result) > 0 {
		return row(result[0]).convert(), nil
	}
	return nil, nil
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteGithubAppByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) createInstall(ctx context.Context, install Install) error {
	_, err := db.Conn(ctx).InsertGithubAppInstall(ctx, pggen.InsertGithubAppInstallParams{
		GithubAppInstallID: sql.String(install.ID),
		InstallID:          pgtype.Int8{Int: install.AppID, Status: pgtype.Present},
		GithubAppID:        sql.String(install.GithubApp.ID),
	})
	return err
}

func (db *pgdb) deleteInstall(ctx context.Context, installID string) error {
	_, err := db.Conn(ctx).DeleteGithubAppInstallByID(ctx, sql.String(installID))
	return err
}
