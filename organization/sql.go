package organization

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a database of state and state versions
type db interface {
	otf.Database

	CreateOrganization(ctx context.Context, org *Organization) error
	GetOrganization(ctx context.Context, name string) (*Organization, error)
	GetOrganizationByID(ctx context.Context, id string) (*Organization, error)
	ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*organizationList, error)
	ListOrganizationsByUser(ctx context.Context, userID string) ([]*Organization, error)
	UpdateOrganization(ctx context.Context, name string, fn func(*Organization) error) (*Organization, error)
	DeleteOrganization(ctx context.Context, name string) error
	GetOrganizationNameByWorkspaceID(ctx context.Context, workspaceID string) (string, error)
}

// pgdb is a database of organizations on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func NewDB(db otf.Database) *pgdb {
	return newPGDB(db)
}

func newPGDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

// CreateOrganization persists an Organization to the DB.
func (db *pgdb) CreateOrganization(ctx context.Context, org *Organization) error {
	_, err := db.InsertOrganization(ctx, pggen.InsertOrganizationParams{
		ID:              sql.String(org.ID()),
		CreatedAt:       sql.Timestamptz(org.CreatedAt()),
		UpdatedAt:       sql.Timestamptz(org.UpdatedAt()),
		Name:            sql.String(org.Name()),
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// UpdateOrganization persists an updated Organization to the DB. The existing
// org is fetched from the DB, the supplied func is invoked on the org, and the
// updated org is persisted back to the DB.
func (db *pgdb) UpdateOrganization(ctx context.Context, name string, fn func(*Organization) error) (*Organization, error) {
	var org *Organization
	err := db.Transaction(ctx, func(tx otf.Database) error {
		result, err := tx.FindOrganizationByNameForUpdate(ctx, sql.String(name))
		if err != nil {
			return err
		}
		org = UnmarshalOrganizationRow(pggen.Organizations(result))
		if err := fn(org); err != nil {
			return err
		}
		_, err = tx.UpdateOrganizationByName(ctx, pggen.UpdateOrganizationByNameParams{
			Name:            sql.String(name),
			NewName:         sql.String(org.Name()),
			SessionRemember: org.SessionRemember(),
			SessionTimeout:  org.SessionTimeout(),
			UpdatedAt:       sql.Timestamptz(org.UpdatedAt()),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return org, err
}

func (db *pgdb) ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*organizationList, error) {
	batch := &pgx.Batch{}

	db.FindOrganizationsBatch(batch, opts.GetLimit(), opts.GetOffset())
	db.CountOrganizationsBatch(batch)
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindOrganizationsScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountOrganizationsScan(results)
	if err != nil {
		return nil, err
	}

	var items []*Organization
	for _, r := range rows {
		items = append(items, UnmarshalOrganizationRow(pggen.Organizations(r)))
	}

	return &organizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *pgdb) ListOrganizationsByUser(ctx context.Context, userID string) ([]*Organization, error) {
	result, err := db.FindOrganizationsByUserID(ctx, sql.String(userID))
	if err != nil {
		return nil, err
	}

	var items []*Organization
	for _, r := range result {
		items = append(items, UnmarshalOrganizationRow(pggen.Organizations(r)))
	}
	return items, nil
}

func (db *pgdb) GetOrganizationNameByWorkspaceID(ctx context.Context, workspaceID string) (string, error) {
	name, err := db.FindOrganizationNameByWorkspaceID(ctx, sql.String(workspaceID))
	if err != nil {
		return "", sql.Error(err)
	}
	return name.String, nil
}

func (db *pgdb) GetOrganization(ctx context.Context, name string) (*Organization, error) {
	r, err := db.FindOrganizationByName(ctx, sql.String(name))
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalOrganizationRow(pggen.Organizations(r)), nil
}

func (db *pgdb) GetOrganizationByID(ctx context.Context, id string) (*Organization, error) {
	r, err := db.FindOrganizationByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalOrganizationRow(pggen.Organizations(r)), nil
}

func (db *pgdb) DeleteOrganization(ctx context.Context, name string) error {
	_, err := db.DeleteOrganizationByName(ctx, sql.String(name))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// UnmarshalOrganizationRow converts an organization database row into an
// organization.
func UnmarshalOrganizationRow(row pggen.Organizations) *Organization {
	return &Organization{
		id:              row.OrganizationID.String,
		createdAt:       row.CreatedAt.Time.UTC(),
		updatedAt:       row.UpdatedAt.Time.UTC(),
		name:            row.Name.String,
		sessionRemember: row.SessionRemember,
		sessionTimeout:  row.SessionTimeout,
	}
}

func UnmarshalOrganizationJSONAPI(model *jsonapi.Organization) *Organization {
	return &Organization{
		id:              model.ExternalID,
		createdAt:       model.CreatedAt,
		name:            model.Name,
		sessionRemember: model.SessionRemember,
		sessionTimeout:  model.SessionTimeout,
	}
}
