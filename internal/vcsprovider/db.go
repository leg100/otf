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
		// provides access to generated SQL queries
		*sql.DB
		// github app service for re-constructing vcs provider from a DB query
		github.GithubAppService
		// map of cloud kind to cloud hostname
		cloudHostnames map[cloud.Kind]string
	}
	// pgRow represents a database row for a vcs provider
	pgRow struct {
		VCSProviderID      pgtype.Text        `json:"vcs_provider_id"`
		Token              pgtype.Text        `json:"token"`
		CreatedAt          pgtype.Timestamptz `json:"created_at"`
		Name               pgtype.Text        `json:"name"`
		Cloud              pgtype.Text        `json:"cloud"`
		OrganizationName   pgtype.Text        `json:"organization_name"`
		GithubAppID        pgtype.Int8        `json:"github_app_id"`
		GithubAppInstallID pgtype.Int8        `json:"github_app_install_id"`
	}
)

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
		Name:             sql.String(provider.Name),
		Cloud:            sql.String(string(provider.Kind)),
		OrganizationName: sql.String(provider.Organization),
		CreatedAt:        sql.Timestamptz(provider.CreatedAt),
		Token:            sql.StringPtr(provider.Token),
	}
	if provider.GithubApp != nil {
		params.GithubAppID = pgtype.Int8{Int: provider.GithubApp.AppCredentials.ID, Status: pgtype.Present}
		params.GithubAppInstallID = pgtype.Int8{Int: provider.GithubApp.ID, Status: pgtype.Present}
	} else {
		params.GithubAppID = pgtype.Int8{Status: pgtype.Null}
		params.GithubAppInstallID = pgtype.Int8{Status: pgtype.Null}
	}
	_, err := db.Conn(ctx).InsertVCSProvider(ctx, params)
	return err
}

func (db *pgdb) update(ctx context.Context, id string, fn func(*VCSProvider) error) error {
	var provider *VCSProvider
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		row, err := q.FindVCSProviderForUpdate(ctx, sql.String(id))
		if err != nil {
			return sql.Error(err)
		}
		provider, err = db.toProvider(ctx, pgRow(row))
		if err != nil {
			return err
		}
		if err := fn(provider); err != nil {
			return err
		}
		_, err = q.UpdateVCSProvider(ctx, pggen.UpdateVCSProviderParams{
			VCSProviderID: sql.String(id),
			Token:         sql.StringPtr(provider.Token),
			Name:          sql.String(provider.Name),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (db *pgdb) get(ctx context.Context, id string) (*VCSProvider, error) {
	row, err := db.Conn(ctx).FindVCSProvider(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.toProvider(ctx, pgRow(row))
}

func (db *pgdb) list(ctx context.Context) ([]*VCSProvider, error) {
	rows, err := db.Conn(ctx).FindVCSProviders(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	var providers []*VCSProvider
	for _, r := range rows {
		provider, err := db.toProvider(ctx, pgRow(r))
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
		provider, err := db.toProvider(ctx, pgRow(r))
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

// unmarshal a vcs provider row from the database.
func (db *pgdb) toProvider(ctx context.Context, row pgRow) (*VCSProvider, error) {
	opts := CreateOptions{
		Organization: row.OrganizationName.String,
		Kind:         cloud.Kind(row.Cloud.String),
		Name:         row.Name.String,
		// GithubAppService: db.Git
	}
	if row.Token.Status == pgtype.Present {
		opts.Token = &row.Token.String
	}
	if row.GithubAppID.Status == pgtype.Present {
		opts.GithubAppInstallID = &row.GithubAppInstallID.Int
	}
	return newProvider(ctx, newOptions{
		CreateOptions:    opts,
		ID:               &row.VCSProviderID.String,
		CreatedAt:        internal.Time(row.CreatedAt.Time.UTC()),
		GithubAppService: db.GithubAppService,
		cloudHostnames:   db.cloudHostnames,
	})
}
