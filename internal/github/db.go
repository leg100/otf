package github

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
)

// appDB is a github app database on postgres
type appDB struct {
	*sql.DB
	baseURL             *internal.WebURL
	skipTLSVerification bool
}

func (db *appDB) create(ctx context.Context, app *App) error {
	_, err := db.Exec(ctx, `
INSERT INTO github_apps (
    github_app_id,
    webhook_secret,
    private_key,
    slug,
    organization
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)`,
		app.ID,
		app.WebhookSecret,
		app.PrivateKey,
		app.Slug,
		app.Organization,
	)
	return err
}

func (db *appDB) get(ctx context.Context) (*App, error) {
	type model struct {
		ID            AppID `db:"github_app_id"`
		Slug          string
		WebhookSecret string  `db:"webhook_secret"`
		PrivateKey    string  `db:"private_key"`
		Organization  *string `db:"organization"`
	}

	rows := db.Query(ctx, `SELECT * FROM github_apps`)
	m, err := sql.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model])
	if err != nil {
		return nil, err
	}

	client, err := NewClient(ClientOptions{
		BaseURL:             db.baseURL,
		SkipTLSVerification: db.skipTLSVerification,
		AppCredentials: &AppCredentials{
			ID:         m.ID,
			PrivateKey: m.PrivateKey,
		},
	})
	if err != nil {
		return nil, err
	}

	return &App{
		ID:            m.ID,
		Slug:          m.Slug,
		WebhookSecret: m.WebhookSecret,
		PrivateKey:    m.PrivateKey,
		Organization:  m.Organization,
		Client:        client,
		GithubURL:     db.baseURL,
	}, nil
}

func (db *appDB) delete(ctx context.Context) error {
	_, err := db.Exec(ctx, `DELETE FROM github_apps `)
	return err
}
