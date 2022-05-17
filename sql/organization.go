package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
)

var (
	_ otf.OrganizationStore = (*OrganizationDB)(nil)
)

type OrganizationDB struct {
	*pgxpool.Pool
}

func NewOrganizationDB(conn *pgxpool.Pool) *OrganizationDB {
	return &OrganizationDB{
		Pool: conn,
	}
}

// Create persists a Organization to the DB.
func (db OrganizationDB) Create(org *otf.Organization) (*otf.Organization, error) {
	q := NewQuerier(db.Pool)
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

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	result, err := q.FindOrganizationByNameForUpdate(ctx, name)
	if err != nil {
		return nil, err
	}
	org, err := otf.UnmarshalOrganizationFromDB(result)
	if err != nil {
		return nil, err
	}

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
		result, err := q.UpdateOrganizationSessionRememberByName(ctx, int32(org.SessionRemember), org.Name)
		if err != nil {
			return nil, err
		}
		org.UpdatedAt = result.UpdatedAt
	}
	if org.SessionTimeout != int(*result.SessionTimeout) {
		result, err := q.UpdateOrganizationSessionTimeoutByName(ctx, int32(org.SessionTimeout), org.Name)
		if err != nil {
			return nil, err
		}
		org.UpdatedAt = result.UpdatedAt
	}

	return org, tx.Commit(ctx)
}

func (db OrganizationDB) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	q := NewQuerier(db.Pool)
	batch := &pgx.Batch{}
	ctx := context.Background()

	q.FindOrganizationsBatch(batch, opts.GetLimit(), opts.GetOffset())
	q.CountOrganizationsBatch(batch)
	results := db.Pool.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindOrganizationsScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountOrganizationsScan(results)
	if err != nil {
		return nil, err
	}
	items, err := otf.UnmarshalOrganizationListFromDB(rows)
	if err != nil {
		return nil, err
	}

	return &otf.OrganizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db OrganizationDB) Get(name string) (*otf.Organization, error) {
	q := NewQuerier(db.Pool)
	ctx := context.Background()

	r, err := q.FindOrganizationByName(ctx, name)
	if err != nil {
		return nil, databaseError(err, findOrganizationByNameSQL)
	}

	return &otf.Organization{
		ID:   r.OrganizationID,
		Name: r.Name,
		Timestamps: otf.Timestamps{
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		SessionRemember: int(r.SessionRemember),
		SessionTimeout:  int(r.SessionTimeout),
	}, nil
}

func (db OrganizationDB) Delete(name string) error {
	q := NewQuerier(db.Pool)
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
