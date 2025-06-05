package vcs

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a VCS provider database on postgres
type pgdb struct {
	// provides access to generated SQL queries
	*sql.DB
	*factory
	schemas map[Kind]ConfigSchema
}

func (db *pgdb) create(ctx context.Context, provider *Provider) error {
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

func (db *pgdb) update(ctx context.Context, id resource.TfeID, fn func(context.Context, *Provider) error) error {
	_, err := sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context) (*Provider, error) {
			rows := db.Query(ctx, `
SELECT *
FROM vcs_providers v
WHERE v.vcs_provider_id = $1
FOR UPDATE OF v
`, id)
			return sql.CollectOneRow(rows, db.scan)
		},
		fn,
		func(ctx context.Context, provider *Provider) error {
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

func (db *pgdb) get(ctx context.Context, id resource.TfeID) (*Provider, error) {
	rows := db.Query(ctx, `
SELECT *
FROM vcs_providers v
WHERE v.vcs_provider_id = $1
`, id)

	return sql.CollectOneRow(rows, db.scan)
}

func (db *pgdb) list(ctx context.Context) ([]*Provider, error) {
	rows := db.Query(ctx, `SELECT * FROM vcs_providers`)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) listByOrganization(ctx context.Context, organization organization.Name) ([]*Provider, error) {
	rows := db.Query(ctx, `
SELECT *
FROM vcs_providers v
WHERE v.organization_name = $1
`, organization)
	return sql.CollectRows(rows, db.scan)
}

func (db *pgdb) listByGithubAppInstall(ctx context.Context, installID int64) ([]*Provider, error) {
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

func (db *pgdb) scan(row pgx.CollectableRow) (*Provider, error) {
	// model represents a database row for a vcs provider
	type model struct {
		VCSProviderID       resource.TfeID `db:"vcs_provider_id"`
		Token               *string
		CreatedAt           time.Time `db:"created_at"`
		Name                string
		VCSKind             Kind              `db:"vcs_kind"`
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
	cfg := Config{
		Token: m.Token,
	}
	if m.InstallID != nil {
		cfg.Installation = &Installation{
			ID:           *m.InstallID,
			AppID:        *m.InstallAppID,
			Username:     m.InstallUsername,
			Organization: m.InstallOrganization,
		}
	}
	schema, ok := db.schemas[m.VCSKind]
	if !ok {
		return nil, errors.New("schema not found")
	}
	client, err := schema.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	provider := Provider{
		ID:           m.VCSProviderID,
		CreatedAt:    m.CreatedAt,
		Organization: m.OrganizationName,
		Name:         m.Name,
		Kind:         m.VCSKind,
		Client:       client,
		Hostname:     schema.Hostname,
	}
	return &provider, nil
}
