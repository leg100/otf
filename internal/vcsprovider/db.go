package vcsprovider

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
)

// pgdb is a VCS provider database on postgres
type pgdb struct {
	// provides access to generated SQL queries
	*sql.DB
	*factory
}

func (db *pgdb) create(ctx context.Context, provider *VCSProvider) error {
	err := db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		args := pgx.NamedArgs{
			"id":           provider.ID,
			"created_at":   provider.CreatedAt,
			"name":         provider.Name,
			"vcs_kind":     provider.Kind,
			"token":        provider.Token,
			"organization": provider.Organization,
		}
		if provider.GithubApp != nil {
			args["github_app_id"] = provider.GithubApp.AppCredentials.ID
		}
		_, err := db.Exec(ctx, `
INSERT INTO vcs_providers (
    vcs_provider_id,
    created_at,
    name,
    vcs_kind,
    token,
    github_app_id,
    organization_name
) VALUES (
	@id,
	@created_at,
	@name,
	@vcs_kind,
	@token,
	@github_app_id,
	@organization
)`, args)
		if err != nil {
			return err
		}
		if provider.GithubApp != nil {
			_, err := db.Exec(ctx, `
INSERT INTO github_app_installs (
    github_app_id,
    install_id,
    username,
    organization,
    vcs_provider_id
) VALUES (
	@id,
	@install_id,
	@username,
	@organization,
	@vcs_provider_id
)`, pgx.NamedArgs{
				"id":              provider.GithubApp.AppCredentials.ID,
				"install_id":      provider.GithubApp.ID,
				"username":        provider.GithubApp.User,
				"organization":    provider.GithubApp.Organization,
				"vcs_provider_id": provider.ID,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (db *pgdb) update(ctx context.Context, id resource.ID, fn func(context.Context, *VCSProvider) error) error {
	_, err := sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*VCSProvider, error) {
			rows := db.Query(ctx, `
SELECT
    v.vcs_provider_id, v.token, v.created_at, v.name, v.vcs_kind, v.organization_name, v.github_app_id,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.vcs_provider_id = $1
FOR UPDATE OF v
`, id)
			return sql.CollectOneRow(rows, db.scan)
		},
		fn,
		func(ctx context.Context, conn sql.Connection, provider *VCSProvider) error {
			_, err := db.Exec(ctx, `
UPDATE vcs_providers
SET name = $1, token = $2
WHERE vcs_provider_id = $3
RETURNING vcs_provider_id, token, created_at, name, vcs_kind, organization_name, github_app_id
`, provider.Name, provider.Token, provider.ID)
			return err
		},
	)
	return err
}

func (db *pgdb) get(ctx context.Context, id resource.ID) (*VCSProvider, error) {
	rows := db.Query(ctx, `
SELECT
    v.vcs_provider_id, v.token, v.created_at, v.name, v.vcs_kind, v.organization_name, v.github_app_id,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.vcs_provider_id = $1
`, id)

	return sql.CollectOneRow(rows, db.scan)
}

func (db *pgdb) list(ctx context.Context) ([]*VCSProvider, error) {
	rows := db.Query(ctx, `
SELECT
    v.vcs_provider_id, v.token, v.created_at, v.name, v.vcs_kind, v.organization_name, v.github_app_id,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
`)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) listByOrganization(ctx context.Context, organization resource.OrganizationName) ([]*VCSProvider, error) {
	rows := db.Query(ctx, `
SELECT
    v.vcs_provider_id, v.token, v.created_at, v.name, v.vcs_kind, v.organization_name, v.github_app_id,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
LEFT JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE v.organization_name = $1
`, organization)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) listByGithubAppInstall(ctx context.Context, installID int64) ([]*VCSProvider, error) {
	rows := db.Query(ctx, `
SELECT
    v.vcs_provider_id, v.token, v.created_at, v.name, v.vcs_kind, v.organization_name, v.github_app_id,
    (ga.*)::"github_apps" AS github_app,
    (gi.*)::"github_app_installs" AS github_app_install
FROM vcs_providers v
JOIN (github_app_installs gi JOIN github_apps ga USING (github_app_id)) USING (vcs_provider_id)
WHERE gi.install_id = $1
`, installID)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) delete(ctx context.Context, id resource.ID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM vcs_providers
WHERE vcs_provider_id = $1
`, id)
	return err
}

type (
	// model represents a database row for a vcs provider
	model struct {
		VCSProviderID    resource.ID `db:"vcs_provider_id"`
		Token            *string
		CreatedAt        time.Time `db:"created_at"`
		Name             string
		VCSKind          vcs.Kind                  `db:"vcs_kind"`
		OrganizationName resource.OrganizationName `db:"organization_name"`
		GithubAppID      *int                      `db:"github_app_id"`
		GithubApp        *githubAppModel           `db:"github_app"`
		GithubAppInstall *githubAppInstallModel    `db:"github_app_install"`
	}

	githubAppModel struct {
		GithubAppID   github.AppID `db:"github_app_id"`
		WebhookSecret string       `db:"webhook_secret"`
		PrivateKey    string       `db:"private_key"`
		Slug          string
		Organization  *string
	}

	githubAppInstallModel struct {
		GithubAppID   int64 `db:"github_app_id"`
		InstallID     int64 `db:"install_id"`
		Username      *string
		Organization  *string
		VCSProviderID string `db:"vcs_provider_id"`
	}
)

func (db *pgdb) scan(row pgx.CollectableRow) (*VCSProvider, error) {
	model, err := pgx.RowToStructByName[model](row)
	if err != nil {
		return nil, err
	}
	opts := CreateOptions{
		Organization: model.OrganizationName,
		Name:         model.Name,
	}
	var creds *github.InstallCredentials
	if model.GithubApp != nil {
		creds = &github.InstallCredentials{
			ID: model.GithubAppInstall.InstallID,
			AppCredentials: github.AppCredentials{
				ID:         model.GithubApp.GithubAppID,
				PrivateKey: model.GithubApp.PrivateKey,
			},
			User:         model.GithubAppInstall.Username,
			Organization: model.GithubAppInstall.Organization,
		}
	} else if model.Token != nil {
		opts.Kind = &model.VCSKind
		opts.Token = model.Token
	}
	provider, err := db.newWithGithubCredentials(opts, creds)
	if err != nil {
		return nil, err
	}
	provider.ID = model.VCSProviderID
	provider.CreatedAt = model.CreatedAt
	return provider, nil
}
