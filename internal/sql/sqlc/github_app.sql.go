// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: github_app.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
)

const deleteGithubApp = `-- name: DeleteGithubApp :one
DELETE
FROM github_apps
WHERE github_app_id = $1
RETURNING github_app_id, webhook_secret, private_key, slug, organization
`

func (q *Queries) DeleteGithubApp(ctx context.Context, githubAppID pgtype.Int8) (GithubApp, error) {
	row := q.db.QueryRow(ctx, deleteGithubApp, githubAppID)
	var i GithubApp
	err := row.Scan(
		&i.GithubAppID,
		&i.WebhookSecret,
		&i.PrivateKey,
		&i.Slug,
		&i.Organization,
	)
	return i, err
}

const findGithubApp = `-- name: FindGithubApp :one
SELECT github_app_id, webhook_secret, private_key, slug, organization
FROM github_apps
`

func (q *Queries) FindGithubApp(ctx context.Context) (GithubApp, error) {
	row := q.db.QueryRow(ctx, findGithubApp)
	var i GithubApp
	err := row.Scan(
		&i.GithubAppID,
		&i.WebhookSecret,
		&i.PrivateKey,
		&i.Slug,
		&i.Organization,
	)
	return i, err
}

const insertGithubApp = `-- name: InsertGithubApp :exec
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
)
`

type InsertGithubAppParams struct {
	GithubAppID   pgtype.Int8
	WebhookSecret pgtype.Text
	PrivateKey    pgtype.Text
	Slug          pgtype.Text
	Organization  pgtype.Text
}

func (q *Queries) InsertGithubApp(ctx context.Context, arg InsertGithubAppParams) error {
	_, err := q.db.Exec(ctx, insertGithubApp,
		arg.GithubAppID,
		arg.WebhookSecret,
		arg.PrivateKey,
		arg.Slug,
		arg.Organization,
	)
	return err
}

const insertGithubAppInstall = `-- name: InsertGithubAppInstall :exec
INSERT INTO github_app_installs (
    github_app_id,
    install_id,
    username,
    organization,
    vcs_provider_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
`

type InsertGithubAppInstallParams struct {
	GithubAppID   pgtype.Int8
	InstallID     pgtype.Int8
	Username      pgtype.Text
	Organization  pgtype.Text
	VCSProviderID resource.ID
}

func (q *Queries) InsertGithubAppInstall(ctx context.Context, arg InsertGithubAppInstallParams) error {
	_, err := q.db.Exec(ctx, insertGithubAppInstall,
		arg.GithubAppID,
		arg.InstallID,
		arg.Username,
		arg.Organization,
		arg.VCSProviderID,
	)
	return err
}
