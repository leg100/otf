package organization

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is a database of organizations on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func newDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

func (db *pgdb) create(ctx context.Context, org *Organization) error {
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
func (db *pgdb) update(ctx context.Context, name string, fn func(*Organization) error) (*Organization, error) {
	var org *Organization
	err := db.Transaction(ctx, func(tx otf.Database) error {
		result, err := tx.FindOrganizationByNameForUpdate(ctx, sql.String(name))
		if err != nil {
			return err
		}
		org = unmarshalRow(pggen.Organizations(result))
		if err := fn(org); err != nil {
			return err
		}
		_, err = tx.UpdateOrganizationByName(ctx, pggen.UpdateOrganizationByNameParams{
			Name:            sql.String(name),
			NewName:         sql.String(org.name),
			SessionRemember: org.sessionRemember,
			SessionTimeout:  org.sessionTimeout,
			UpdatedAt:       sql.Timestamptz(org.updatedAt),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return org, err
}

func (db *pgdb) list(ctx context.Context, opts ListOptions) (*OrganizationList, error) {
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
		items = append(items, unmarshalRow(pggen.Organizations(r)))
	}

	return &OrganizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *pgdb) listByUser(ctx context.Context, userID string, opts ListOptions) (*OrganizationList, error) {
	batch := &pgx.Batch{}

	db.FindOrganizationsByUserIDBatch(batch, pggen.FindOrganizationsByUserIDParams{
		UserID: sql.String(userID),
		Limit:  opts.GetLimit(),
		Offset: opts.GetOffset(),
	})
	db.CountOrganizationsByUserIDBatch(batch, sql.String(userID))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindOrganizationsByUserIDScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountOrganizationsByUserIDScan(results)
	if err != nil {
		return nil, err
	}

	var items []*Organization
	for _, r := range rows {
		items = append(items, unmarshalRow(pggen.Organizations(r)))
	}

	return &OrganizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *pgdb) get(ctx context.Context, name string) (*Organization, error) {
	r, err := db.FindOrganizationByName(ctx, sql.String(name))
	if err != nil {
		return nil, sql.Error(err)
	}
	return unmarshalRow(pggen.Organizations(r)), nil
}

func (db *pgdb) getByID(ctx context.Context, id string) (*Organization, error) {
	r, err := db.FindOrganizationByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return unmarshalRow(pggen.Organizations(r)), nil
}

func (db *pgdb) delete(ctx context.Context, name string) error {
	_, err := db.DeleteOrganizationByName(ctx, sql.String(name))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// unmarshalRow converts an organization database row into an
// organization.
func unmarshalRow(row pggen.Organizations) *Organization {
	return &Organization{
		id:              row.OrganizationID.String,
		createdAt:       row.CreatedAt.Time.UTC(),
		updatedAt:       row.UpdatedAt.Time.UTC(),
		name:            row.Name.String,
		sessionRemember: row.SessionRemember,
		sessionTimeout:  row.SessionTimeout,
	}
}
