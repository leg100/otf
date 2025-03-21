package vcsprovider

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
)

var q = &Queries{}

type (
	// pgdb is a VCS provider database on postgres
	pgdb struct {
		// provides access to generated SQL queries
		*sql.DB
		*factory
	}
	// pgRow represents a database row for a vcs provider
	pgRow struct {
		VCSProviderID    resource.TfeID
		Token            pgtype.Text
		CreatedAt        pgtype.Timestamptz
		Name             pgtype.Text
		VCSKind          pgtype.Text
		OrganizationName resource.OrganizationName
		GithubAppID      pgtype.Int8
		GithubApp        *GithubApp
		GithubAppInstall *GithubAppInstall
	}
)

func (db *pgdb) create(ctx context.Context, provider *VCSProvider) error {
	err := db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		params := InsertVCSProviderParams{
			VCSProviderID:    provider.ID,
			Name:             sql.String(provider.Name),
			VCSKind:          sql.String(string(provider.Kind)),
			OrganizationName: provider.Organization,
			CreatedAt:        sql.Timestamptz(provider.CreatedAt),
			Token:            sql.StringPtr(provider.Token),
		}
		if provider.GithubApp != nil {
			params.GithubAppID = pgtype.Int8{Int64: provider.GithubApp.AppCredentials.ID, Valid: true}
		}
		if err := q.InsertVCSProvider(ctx, conn, params); err != nil {
			return err
		}
		if provider.GithubApp != nil {
			err := (&github.Queries{}).InsertGithubAppInstall(ctx, conn, github.InsertGithubAppInstallParams{
				GithubAppID:   pgtype.Int8{Int64: provider.GithubApp.AppCredentials.ID, Valid: true},
				InstallID:     pgtype.Int8{Int64: provider.GithubApp.ID, Valid: true},
				Username:      sql.StringPtr(provider.GithubApp.User),
				Organization:  sql.StringPtr(provider.GithubApp.Organization),
				VCSProviderID: provider.ID,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (db *pgdb) update(ctx context.Context, id resource.TfeID, fn func(context.Context, *VCSProvider) error) error {
	_, err := sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*VCSProvider, error) {
			row, err := q.FindVCSProviderForUpdate(ctx, conn, id)
			if err != nil {
				return nil, err
			}
			return db.toProvider(ctx, pgRow(row))
		},
		fn,
		func(ctx context.Context, conn sql.Connection, provider *VCSProvider) error {
			_, err := q.UpdateVCSProvider(ctx, conn, UpdateVCSProviderParams{
				VCSProviderID: id,
				Token:         sql.StringPtr(provider.Token),
				Name:          sql.String(provider.Name),
			})
			return err
		},
	)
	return err
}

func (db *pgdb) get(ctx context.Context, id resource.TfeID) (*VCSProvider, error) {
	row, err := q.FindVCSProvider(ctx, db.Conn(ctx), id)
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.toProvider(ctx, pgRow(row))
}

func (db *pgdb) list(ctx context.Context) ([]*VCSProvider, error) {
	rows, err := q.FindVCSProviders(ctx, db.Conn(ctx))
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

func (db *pgdb) listByOrganization(ctx context.Context, organization resource.OrganizationName) ([]*VCSProvider, error) {
	rows, err := q.FindVCSProvidersByOrganization(ctx, db.Conn(ctx), organization)
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
	rows, err := q.FindVCSProvidersByGithubAppInstallID(ctx, db.Conn(ctx), sql.Int8(int(installID)))
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

func (db *pgdb) delete(ctx context.Context, id resource.TfeID) error {
	_, err := q.DeleteVCSProviderByID(ctx, db.Conn(ctx), id)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// unmarshal a vcs provider row from the database.
func (db *pgdb) toProvider(ctx context.Context, row pgRow) (*VCSProvider, error) {
	opts := CreateOptions{
		Organization: row.OrganizationName,
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
	return db.fromDB(ctx, opts, creds, row.VCSProviderID, row.CreatedAt.Time.UTC())
}
