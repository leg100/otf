package vcsprovider

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/vcs"
)

type (
	// pgdb is a VCS provider database on postgres
	pgdb struct {
		// provides access to generated SQL queries
		*sql.DB
		*factory
	}
	// pgRow represents a database row for a vcs provider
	pgRow struct {
		VCSProviderID    pgtype.Text
		Token            pgtype.Text
		CreatedAt        pgtype.Timestamptz
		Name             pgtype.Text
		VCSKind          pgtype.Text
		OrganizationName pgtype.Text
		GithubAppID      pgtype.Int8
		GithubApp        *sqlc.GithubApp
		GithubAppInstall *sqlc.GithubAppInstall
	}
)

func (db *pgdb) create(ctx context.Context, provider *VCSProvider) error {
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		params := sqlc.InsertVCSProviderParams{
			VCSProviderID:    sql.String(provider.ID.String()),
			Name:             sql.String(provider.Name),
			VCSKind:          sql.String(string(provider.Kind)),
			OrganizationName: sql.String(provider.Organization),
			CreatedAt:        sql.Timestamptz(provider.CreatedAt),
			Token:            sql.StringPtr(provider.Token),
		}
		if provider.GithubApp != nil {
			params.GithubAppID = pgtype.Int8{Int64: provider.GithubApp.AppCredentials.ID, Valid: true}
		}
		if err := db.Querier(ctx).InsertVCSProvider(ctx, params); err != nil {
			return err
		}
		if provider.GithubApp != nil {
			err := db.Querier(ctx).InsertGithubAppInstall(ctx, sqlc.InsertGithubAppInstallParams{
				GithubAppID:   pgtype.Int8{Int64: provider.GithubApp.AppCredentials.ID, Valid: true},
				InstallID:     pgtype.Int8{Int64: provider.GithubApp.ID, Valid: true},
				Username:      sql.StringPtr(provider.GithubApp.User),
				Organization:  sql.StringPtr(provider.GithubApp.Organization),
				VCSProviderID: sql.String(provider.ID.String()),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (db *pgdb) update(ctx context.Context, id resource.ID, fn func(*VCSProvider) error) error {
	var provider *VCSProvider
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
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
		_, err = q.UpdateVCSProvider(ctx, sqlc.UpdateVCSProviderParams{
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

func (db *pgdb) get(ctx context.Context, id resource.ID) (*VCSProvider, error) {
	row, err := db.Querier(ctx).FindVCSProvider(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.toProvider(ctx, pgRow(row))
}

func (db *pgdb) list(ctx context.Context) ([]*VCSProvider, error) {
	rows, err := db.Querier(ctx).FindVCSProviders(ctx)
	if err != nil {
		return nil, sql.Error(err)
	}
	providers := make([]*VCSProvider, len(rows))
	for i, r := range rows {
		provider, err := db.toProvider(ctx, pgRow(r))
		if err != nil {
			return nil, err
		}
		providers[i] = provider
	}
	return providers, nil
}

func (db *pgdb) listByOrganization(ctx context.Context, organization string) ([]*VCSProvider, error) {
	rows, err := db.Querier(ctx).FindVCSProvidersByOrganization(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	providers := make([]*VCSProvider, len(rows))
	for i, r := range rows {
		provider, err := db.toProvider(ctx, pgRow(r))
		if err != nil {
			return nil, err
		}
		providers[i] = provider
	}
	return providers, nil
}

func (db *pgdb) listByGithubAppInstall(ctx context.Context, installID int64) ([]*VCSProvider, error) {
	rows, err := db.Querier(ctx).FindVCSProvidersByGithubAppInstallID(ctx,
		sql.Int8(int(installID)),
	)
	if err != nil {
		return nil, sql.Error(err)
	}
	providers := make([]*VCSProvider, len(rows))
	for i, r := range rows {
		provider, err := db.toProvider(ctx, pgRow(r))
		if err != nil {
			return nil, err
		}
		providers[i] = provider
	}
	return providers, nil
}

func (db *pgdb) delete(ctx context.Context, id resource.ID) error {
	_, err := db.Querier(ctx).DeleteVCSProviderByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// unmarshal a vcs provider row from the database.
func (db *pgdb) toProvider(ctx context.Context, row pgRow) (*VCSProvider, error) {
	opts := CreateOptions{
		Organization: row.OrganizationName.String,
		Name:         row.Name.String,
		// GithubAppService: db.Git
	}
	if row.Token.Valid {
		opts.Token = &row.Token.String
		kind := vcs.Kind(row.VCSKind.String)
		opts.Kind = &kind
	}
	var creds *github.InstallCredentials
	if row.GithubApp != nil {
		creds = &github.InstallCredentials{
			ID: row.GithubAppInstall.InstallID.Int64,
			AppCredentials: github.AppCredentials{
				ID:         row.GithubApp.GithubAppID.Int64,
				PrivateKey: row.GithubApp.PrivateKey.String,
			},
		}
		if row.GithubAppInstall.Username.Valid {
			creds.User = &row.GithubAppInstall.Username.String
		}
		if row.GithubAppInstall.Organization.Valid {
			creds.Organization = &row.GithubAppInstall.Organization.String
		}
	}
	return db.fromDB(ctx, opts, creds, row.VCSProviderID.String, row.CreatedAt.Time.UTC())
}
