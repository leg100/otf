package vcsprovider

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is a VCS provider database on postgres
type pgdb struct {
	otf.DB // provides access to generated SQL queries
	*factory
}

func newDB(db otf.DB, cloudService cloud.Service) *pgdb {
	return &pgdb{db, &factory{cloudService}}
}

// CreateVCSProvider inserts an agent token, associating it with an organization
func (db *pgdb) create(ctx context.Context, provider *otf.VCSProvider) error {
	_, err := db.InsertVCSProvider(ctx, pggen.InsertVCSProviderParams{
		VCSProviderID:    sql.String(provider.ID),
		Token:            sql.String(provider.Token),
		Name:             sql.String(provider.Name),
		Cloud:            sql.String(provider.CloudConfig.Name),
		OrganizationName: sql.String(provider.Organization),
		CreatedAt:        sql.Timestamptz(provider.CreatedAt),
	})
	return err
}

func (db *pgdb) get(ctx context.Context, id string) (*otf.VCSProvider, error) {
	row, err := db.FindVCSProvider(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.unmarshal(pgRow(row))
}

func (db *pgdb) list(ctx context.Context, organization string) ([]*otf.VCSProvider, error) {
	rows, err := db.FindVCSProviders(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	var providers []*otf.VCSProvider
	for _, r := range rows {
		provider, err := db.unmarshal(pgRow(r))
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

// DeleteVCSProvider deletes an agent token.
func (db *pgdb) delete(ctx context.Context, id string) error {
	_, err := db.DeleteVCSProviderByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// pgRow represents a database row for a vcs provider
type pgRow struct {
	VCSProviderID    pgtype.Text        `json:"id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Name             pgtype.Text        `json:"name"`
	Cloud            pgtype.Text        `json:"cloud"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

// UnmarshalVCSProviderRow unmarshals a vcs provider row from the database.
func (db *pgdb) unmarshal(row pgRow) (*otf.VCSProvider, error) {
	return db.new(createOptions{
		ID:           &row.VCSProviderID.String,
		CreatedAt:    otf.Time(row.CreatedAt.Time.UTC()),
		Organization: row.OrganizationName.String,
		Token:        row.Token.String,
		Name:         row.Name.String,
		Cloud:        row.Cloud.String,
	})
}
