package vcs

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a VCS provider database on postgres
type pgdb struct {
	// provides access to generated SQL queries
	*sql.DB
	kinds *kindDB
}

func (db *pgdb) create(ctx context.Context, provider *Provider) error {
	args := pgx.NamedArgs{
		"id":                   provider.ID,
		"token":                provider.Token,
		"created_at":           provider.CreatedAt,
		"name":                 provider.Name,
		"vcs_kind":             provider.Kind.ID,
		"organization":         provider.Organization,
		"base_url":             provider.BaseURL,
		"api_url":              provider.apiURL,
		"tfe_service_provider": provider.serviceProviderType,
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
    organization_name,
    install_app_id,
    install_id,
    install_username,
    install_organization,
	base_url,
	api_url,
	tfe_service_provider
) VALUES (
	@id,
	@token,
	@created_at,
	@name,
	@vcs_kind,
	@organization,
    @install_app_id,
    @install_id,
    @install_username,
    @install_organization,
	@base_url,
	@api_url,
	@tfe_service_provider
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
			return db.scanOne(ctx, rows)
		},
		fn,
		func(ctx context.Context, provider *Provider) error {
			args := pgx.NamedArgs{
				"id":       provider.ID,
				"token":    provider.Token,
				"name":     provider.Name,
				"base_url": provider.BaseURL,
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
	base_url = @base_url,
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

	return db.scanOne(ctx, rows)
}

func (db *pgdb) listByOrganization(ctx context.Context, organization organization.Name) ([]*Provider, error) {
	rows := db.Query(ctx, `
SELECT *
FROM vcs_providers v
WHERE v.organization_name = $1
ORDER BY v.name
`, organization)
	return db.scanMany(ctx, rows)
}

func (db *pgdb) listByInstall(ctx context.Context, installID int64) ([]*Provider, error) {
	rows := db.Query(ctx, `
SELECT *
FROM vcs_providers v
WHERE v.install_id = $1
ORDER BY v.name
`, installID)
	return db.scanMany(ctx, rows)
}

func (db *pgdb) delete(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM vcs_providers
WHERE vcs_provider_id = $1
`, id)
	return err
}

// model represents a database row for a vcs provider
type model struct {
	VCSProviderID       resource.TfeID `db:"vcs_provider_id"`
	Token               *string
	CreatedAt           time.Time `db:"created_at"`
	Name                string
	VCSKind             KindID                 `db:"vcs_kind"`
	OrganizationName    organization.Name      `db:"organization_name"`
	InstallAppID        *int64                 `db:"install_app_id"`
	InstallID           *int64                 `db:"install_id"`
	InstallUsername     *string                `db:"install_username"`
	InstallOrganization *string                `db:"install_organization"`
	BaseURL             *internal.WebURL       `db:"base_url"`
	APIURL              *internal.WebURL       `db:"api_url"`
	TFEServiceProvider  TFEServiceProviderType `db:"tfe_service_provider"`
}

func (db *pgdb) scanOne(ctx context.Context, row pgx.Rows) (*Provider, error) {
	model, err := sql.CollectOneRow(row, pgx.RowToStructByName[model])
	if err != nil {
		return nil, err
	}
	return db.toProvider(ctx, model)
}

func (db *pgdb) scanMany(ctx context.Context, row pgx.Rows) ([]*Provider, error) {
	// convert to models
	models, err := sql.CollectRows(row, pgx.RowToStructByName[model])
	if err != nil {
		return nil, err
	}
	// convert models to providers
	providers := make([]*Provider, len(models))
	for i, m := range models {
		providers[i], err = db.toProvider(ctx, m)
		if err != nil {
			return nil, err
		}
	}
	return providers, nil
}

func (db *pgdb) toProvider(ctx context.Context, m model) (*Provider, error) {
	cfg := ClientConfig{
		Token:   m.Token,
		BaseURL: m.BaseURL,
	}
	if m.InstallID != nil {
		cfg.Installation = &Installation{
			ID:           *m.InstallID,
			AppID:        *m.InstallAppID,
			Username:     m.InstallUsername,
			Organization: m.InstallOrganization,
		}
	}
	kind, err := db.kinds.GetKind(m.VCSKind)
	if err != nil {
		return nil, err
	}
	client, err := kind.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	provider := Provider{
		ID:                  m.VCSProviderID,
		CreatedAt:           m.CreatedAt,
		Organization:        m.OrganizationName,
		Name:                m.Name,
		Kind:                kind,
		Client:              client,
		Token:               m.Token,
		Installation:        cfg.Installation,
		BaseURL:             m.BaseURL,
		apiURL:              m.APIURL,
		serviceProviderType: m.TFEServiceProvider,
	}
	return &provider, nil
}
