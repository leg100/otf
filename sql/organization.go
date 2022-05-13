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
		ID:              org.ID,
		Name:            org.Name,
		SessionRemember: int32(org.SessionRemember),
		SessionTimeout:  int32(org.SessionTimeout),
	})
	if err != nil {
		return nil, databaseError(err, insertOrganizationSQL)
	}

	org.CreatedAt = result.CreatedAt
	org.UpdatedAt = result.UpdatedAt

	return org, nil
}

// Update persists an updated Organization to the DB. The existing org is
// fetched from the DB, the supplied func is invoked on the org, and the updated
// org is persisted back to the DB.
func (db OrganizationDB) Update(name string, fn func(*otf.Organization) error) (*otf.Organization, error) {
	ctx := context.Background()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	result, err := q.FindOrganizationByNameForUpdate(ctx, name)
	if err != nil {
		return nil, err
	}
	org := convertOrganizationComposite(Organizations(result))

	if err := fn(org); err != nil {
		return nil, err
	}

	if org.Name != *result.Name {
		result, err := q.UpdateOrganizationNameByName(ctx, org.Name, name)
		if err != nil {
			return nil, err
		}
		org.UpdatedAt = result.UpdatedAt
	}
	if org.SessionRemember != int(*result.SessionRemember) {
		result, err := q.UpdateOrganizationSessionRememberByName(ctx, int32(org.SessionRemember), name)
		if err != nil {
			return nil, err
		}
		org.UpdatedAt = result.UpdatedAt
	}
	if org.SessionTimeout != int(*result.SessionTimeout) {
		result, err := q.UpdateOrganizationSessionTimeoutByName(ctx, int32(org.SessionTimeout), name)
		if err != nil {
			return nil, err
		}
		org.UpdatedAt = result.UpdatedAt
	}

	return org, tx.Commit(ctx)
}

func (db OrganizationDB) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.FindOrganizations(ctx, opts.GetLimit(), opts.GetOffset())
	if err != nil {
		return nil, err
	}

	var items []*otf.Organization
	for _, r := range result {
		items = append(items, &otf.Organization{
			Name: *r.Name,
			Timestamps: otf.Timestamps{
				CreatedAt: r.CreatedAt,
				UpdatedAt: r.UpdatedAt,
			},
			SessionRemember: int(*r.SessionRemember),
			SessionTimeout:  int(*r.SessionTimeout),
		})
	}

	return &otf.OrganizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, getCount(result)),
	}, nil
}

func (db OrganizationDB) Get(name string) (*otf.Organization, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	r, err := q.FindOrganizationByName(ctx, name)
	if err != nil {
		return nil, databaseError(err, findOrganizationByNameSQL)
	}

	return &otf.Organization{
		Name: r.Name,
		Timestamps: otf.Timestamps{
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		SessionRemember: int(*r.SessionRemember),
		SessionTimeout:  int(*r.SessionTimeout),
	}, nil
}

func (db OrganizationDB) Delete(name string) error {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.DeleteOrganization(ctx, name)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
