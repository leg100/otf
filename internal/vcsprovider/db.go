package vcsprovider

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/organization"
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
	args := pgx.NamedArgs{
		"id":           provider.ID,
		"token":        provider.Token,
		"created_at":   provider.CreatedAt,
		"name":         provider.Name,
		"vcs_kind":     provider.Kind,
		"organization": provider.Organization,
	}
	if provider.Installation != nil {
		args["install_app_id"] = provider.Installation.AppID
		args["install_id"] = provider.Installation.ID
		args["install_username"] = provider.Installation.Username
		args["install_organization"] = provider.Installation.Organization
	}

	_, err := db.Exec(ctx, `
INSERT INTO vcs_providers (
    vcs_provider_id,
    token,
    created_at,
    name,
    vcs_kind,
    organization_name
    install_app_id,
    install_id,
    install_username,
    install_organization,
) VALUES (
	@id,
	@token,
	@created_at,
	@name,
	@vcs_kind,
	@organization
    @install_app_id,
    @install_id,
    @install_username,
    @install_organization,
)`, args)
	return err
}

func (db *pgdb) update(ctx context.Context, id resource.TfeID, fn func(context.Context, *VCSProvider) error) error {
	_, err := sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context) (*VCSProvider, error) {
			rows := db.Query(ctx, `
SELECT *
FROM vcs_providers v
WHERE v.vcs_provider_id = $1
FOR UPDATE OF v
`, id)
			return sql.CollectOneRow(rows, db.scan)
		},
		fn,
		func(ctx context.Context, provider *VCSProvider) error {
			args := pgx.NamedArgs{
				"id":    provider.ID,
				"token": provider.Token,
				"name":  provider.Name,
			}
			if provider.Installation != nil {
				args["install_app_id"] = provider.Installation.AppID
				args["install_id"] = provider.Installation.ID
				args["install_username"] = provider.Installation.Username
				args["install_organization"] = provider.Installation.Organization
			}
			_, err := db.Exec(ctx, `
UPDATE vcs_providers
SET
	name = @name,
	token = @token,
	install_app_id = @install_app_id,
	install_id = @install_id,
	install_username = @install_username,
	install_organization = @install_organization
WHERE vcs_provider_id = @id
`, args)
			return err
		},
	)
	return err
}

func (db *pgdb) get(ctx context.Context, id resource.TfeID) (*VCSProvider, error) {
	rows := db.Query(ctx, `
SELECT *
FROM vcs_providers v
WHERE v.vcs_provider_id = $1
`, id)

	return sql.CollectOneRow(rows, db.scan)
}

func (db *pgdb) list(ctx context.Context) ([]*VCSProvider, error) {
	rows := db.Query(ctx, `SELECT * FROM vcs_providers`)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) listByOrganization(ctx context.Context, organization organization.Name) ([]*VCSProvider, error) {
	rows := db.Query(ctx, `
SELECT *
FROM vcs_providers v
WHERE v.organization_name = $1
`, organization)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) listByGithubAppInstall(ctx context.Context, installID int64) ([]*VCSProvider, error) {
	rows := db.Query(ctx, `
SELECT *
FROM vcs_providers v
WHERE gi.install_id = $1
`, installID)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) delete(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM vcs_providers
WHERE vcs_provider_id = $1
`, id)
	return err
}

func (db *pgdb) scan(row pgx.CollectableRow) (*VCSProvider, error) {
	// model represents a database row for a vcs provider
	type model struct {
		VCSProviderID       resource.TfeID `db:"vcs_provider_id"`
		Token               *string
		CreatedAt           time.Time `db:"created_at"`
		Name                string
		VCSKind             vcs.Kind          `db:"vcs_kind"`
		OrganizationName    organization.Name `db:"organization_name"`
		InstallAppID        *int64            `db:"install_app_id"`
		InstallID           *int64            `db:"install_id"`
		InstallUsername     *string           `db:"install_username"`
		InstallOrganization *string           `db:"install_organization"`
	}

	m, err := pgx.RowToStructByName[model](row)
	if err != nil {
		return nil, err
	}
	opts := CreateOptions{
		Organization: m.OrganizationName,
		Name:         m.Name,
		Kind:         m.VCSKind,
		Config: Config{
			Token: m.Token,
		},
	}
	if m.InstallID != nil {
		opts.Installation = &Installation{
			ID:           *m.InstallID,
			AppID:        *m.InstallAppID,
			Username:     m.InstallUsername,
			Organization: m.InstallOrganization,
		}
	}
	provider, err := db.newProvider(opts)
	if err != nil {
		return nil, err
	}
	provider.ID = m.VCSProviderID
	provider.CreatedAt = m.CreatedAt
	return provider, nil
}
