package github

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a github app database on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) create(ctx context.Context, app *App) error {
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

func (db *pgdb) get(ctx context.Context) (*App, error) {
	rows := db.Query(ctx, `SELECT * FROM github_apps`)
	return sql.CollectOneRow(rows, pgx.RowToAddrOfStructByName[App])
}

func (db *pgdb) delete(ctx context.Context) error {
	return db.Lock(ctx, "github_apps", func(ctx context.Context, conn sql.Connection) error {
		rows := db.Query(ctx, `SELECT * FROM github_apps`)
		result, err := sql.CollectOneRow(rows, pgx.RowToAddrOfStructByName[App])
		if err != nil {
			return err
		}
		_, err = db.Exec(ctx, `
DELETE
FROM github_apps
WHERE github_app_id = $1
`, result.ID)
		return err
	})
}
