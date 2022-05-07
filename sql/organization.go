package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var (
	_ otf.OrganizationStore = (*OrganizationDB)(nil)
)

type OrganizationDB struct {
	*pgx.Conn
}

type organizationRow interface {
	GetOrganizationID() *string
	GetName() *string
	GetSessionTimeout() *int32
	GetSessionRemember() *int32

	Timestamps
}

type organizationListResult interface {
	GetOrganizations() []Organizations

	GetFullCount() *int
}

func NewOrganizationDB(conn *pgx.Conn) *OrganizationDB {
	return &OrganizationDB{
		Conn: conn,
	}
}

// Create persists a Organization to the DB.
func (db OrganizationDB) Create(org *otf.Organization) (*otf.Organization, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.InsertOrganization(ctx, InsertOrganizationParams{
		ID:              &org.ID,
		Name:            &org.Name,
		SessionRemember: int32(org.SessionRemember),
		SessionTimeout:  int32(org.SessionTimeout),
	})
	if err != nil {
		return nil, err
	}

	return convertOrganization(result), nil
}

// Update persists an updated Organization to the DB. The existing org is
// fetched from the DB, the supplied func is invoked on the org, and the updated
// org is persisted back to the DB.
func (db OrganizationDB) Update(name string, opts otf.OrganizationUpdateOptions) (*otf.Organization, error) {
	ctx := context.Background()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	var modified bool

	if opts.Name != nil {
		_, err := q.UpdateOrganizationNameByName(ctx, opts.Name, &name)
		if err != nil {
			return nil, err
		}
		modified = true
	}

	if opts.SessionTimeout != nil {
		_, err := q.UpdateOrganizationSessionTimeoutByName(ctx, int32(*opts.SessionTimeout), &name)
		if err != nil {
			return nil, err
		}
		modified = true
	}

	if opts.SessionRemember != nil {
		_, err := q.UpdateOrganizationSessionRememberByName(ctx, int32(*opts.SessionRemember), &name)
		if err != nil {
			return nil, err
		}
		modified = true
	}

	if modified {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
	}

	return getOrganization(ctx, q, name)
}

func (db OrganizationDB) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.FindOrganizations(ctx, opts.GetLimit(), opts.GetOffset())
	if err != nil {
		return nil, err
	}

	var items []*otf.Organization
	for _, r := range result.Organizations {
		items = append(items, convertOrganization(r))
	}

	return &otf.OrganizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *result.FullCount),
	}, nil
}

func (db OrganizationDB) Get(name string) (*otf.Organization, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	return getOrganization(ctx, q, name)
}

func (db OrganizationDB) Delete(name string) error {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.DeleteOrganization(ctx, &name)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func getOrganization(ctx context.Context, q *DBQuerier, name string) (*otf.Organization, error) {
	result, err := q.FindOrganizationByName(ctx, &name)
	if err != nil {
		return nil, err
	}

	return convertOrganization(result), nil
}

func convertOrganization(row organizationRow) *otf.Organization {
	organization := otf.Organization{
		ID:              *row.GetOrganizationID(),
		Name:            *row.GetName(),
		Timestamps:      convertTimestamps(row),
		SessionRemember: int(*row.GetSessionRemember()),
		SessionTimeout:  int(*row.GetSessionTimeout()),
	}

	return &organization
}
