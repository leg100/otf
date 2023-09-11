package vcsprovider

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is a VCS provider database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
		CloudService
	}
	// pgRow represents a database row for a vcs provider
	pgRow struct {
		VCSProviderID    pgtype.Text        `json:"id"`
		Token            pgtype.Text        `json:"token"`
		CreatedAt        pgtype.Timestamptz `json:"created_at"`
		Name             pgtype.Text        `json:"name"`
		Cloud            pgtype.Text        `json:"cloud"`
		OrganizationName pgtype.Text        `json:"organization_name"`
	}
)

func newDB(db *sql.DB, cloudService cloud.Service) *pgdb {
	return &pgdb{db, cloudService}
}

// GetByID implements pubsub.Getter
func (db *pgdb) GetByID(ctx context.Context, providerID string, action pubsub.DBAction) (any, error) {
	if action == pubsub.DeleteDBAction {
		return &VCSProvider{ID: providerID}, nil
	}
	return db.get(ctx, providerID)
}

func (db *pgdb) create(ctx context.Context, provider *VCSProvider) error {
	_, err := db.Conn(ctx).InsertVCSProvider(ctx, pggen.InsertVCSProviderParams{
		VCSProviderID:    sql.String(provider.ID),
		Token:            sql.String(provider.Token),
		Name:             sql.String(provider.Name),
		Cloud:            sql.String(provider.CloudConfig.Name),
		OrganizationName: sql.String(provider.Organization),
		CreatedAt:        sql.Timestamptz(provider.CreatedAt),
	})
	return err
}

func (db *pgdb) update(ctx context.Context, id string, fn func(*VCSProvider) error) error {
	var provider *VCSProvider
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		row, err := q.FindVCSProviderForUpdate(ctx, sql.String(id))
		if err != nil {
			return sql.Error(err)
		}
		provider, err = db.unmarshal(pgRow(row))
		if err != nil {
			return err
		}
		if err := fn(provider); err != nil {
			return err
		}
		_, err = q.UpdateVCSProvider(ctx, pggen.UpdateVCSProviderParams{
			VCSProviderID: sql.String(id),
			Token:         sql.String(provider.Token),
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

// unmarshal a vcs provider row from the database.
func (db *pgdb) unmarshal(row pgRow) (*VCSProvider, error) {
	opts := CreateOptions{
		ID:           &row.VCSProviderID.String,
		CreatedAt:    internal.Time(row.CreatedAt.Time.UTC()),
		Organization: row.OrganizationName.String,
		Token:        row.Token.String,
		Cloud:        row.Cloud.String,
		Name:         row.Name.String,
	}
	return newProvider(db.CloudService, opts)
}
