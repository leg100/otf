package vcsprovider

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is a VCS provider database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
		*factory
	}
	// pgRow represents a database row for a vcs provider
	pgRow struct {
		GithubAppID      pgtype.Int8        `json:"github_app_id"`
		VCSProviderID    pgtype.Text        `json:"vcs_provider_id"`
		Token            pgtype.Text        `json:"token"`
		CreatedAt        pgtype.Timestamptz `json:"created_at"`
		Name             pgtype.Text        `json:"name"`
		Cloud            pgtype.Text        `json:"cloud"`
		OrganizationName pgtype.Text        `json:"organization_name"`
		WebhookSecret    pgtype.Text        `json:"webhook_secret"`
		PrivateKey       pgtype.Text        `json:"private_key"`
	}
)

func newDB(db *sql.DB, cloudService cloud.Service) *pgdb {
	return &pgdb{db, &factory{cloudService}}
}

// unmarshal a vcs provider row from the database.
func (db *pgdb) unmarshal(row pgRow) (*VCSProvider, error) {
	opts := CreateOptions{
		ID:           &row.VCSProviderID.String,
		CreatedAt:    internal.Time(row.CreatedAt.Time.UTC()),
		Organization: row.OrganizationName.String,
		Name:         row.Name.String,
		Cloud:        row.Cloud.String,
	}
	if row.Token.Status == pgtype.Present {
		opts.Token = &row.Token.String
	}
	if row.GithubAppID.Status == pgtype.Present {
		opts.GithubApp = &github.Install{
			AppID:         row.GithubAppID.Int,
			WebhookSecret: row.WebhookSecret.String,
			PrivateKey:    row.PrivateKey.String,
		}
	}
	return db.new(opts)
}

// GetByID implements pubsub.Getter
func (db *pgdb) GetByID(ctx context.Context, providerID string, action pubsub.DBAction) (any, error) {
	if action == pubsub.DeleteDBAction {
		return &VCSProvider{ID: providerID}, nil
	}
	return db.get(ctx, providerID)
}

func (db *pgdb) create(ctx context.Context, provider *VCSProvider) error {
	params := pggen.InsertVCSProviderParams{
		VCSProviderID:    sql.String(provider.ID),
		Token:            sql.StringPtr(provider.Token),
		Name:             sql.String(provider.Name),
		Cloud:            sql.String(provider.CloudConfig.Name),
		OrganizationName: sql.String(provider.Organization),
		CreatedAt:        sql.Timestamptz(provider.CreatedAt),
	}
	if provider.GithubApp != nil {
		params.GithubAppID = pgtype.Int8{Int: provider.GithubApp.AppID, Status: pgtype.Null}
	} else {
		params.GithubAppID.Status = pgtype.Null
	}
	_, err := db.Conn(ctx).InsertVCSProvider(ctx, params)
	return err
}

func (db *pgdb) get(ctx context.Context, id string) (*VCSProvider, error) {
	row, err := db.Conn(ctx).FindVCSProvider(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(pgRow(row))
}

func (db *pgdb) list(ctx context.Context) ([]*VCSProvider, error) {
	rows, err := db.Conn(ctx).FindVCSProviders(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	var providers []*VCSProvider
	for _, r := range rows {
		provider, err := db.unmarshal(pgRow(r))
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func (db *pgdb) listByOrganization(ctx context.Context, organization string) ([]*VCSProvider, error) {
	rows, err := db.Conn(ctx).FindVCSProvidersByOrganization(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	var providers []*VCSProvider
	for _, r := range rows {
		provider, err := db.unmarshal(pgRow(r))
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func (db *pgdb) delete(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteVCSProviderByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
