package sql

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
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

// Create persists an Organization to the DB.
func (db OrganizationDB) Create(org *otf.Organization) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()
	_, err := q.InsertOrganization(ctx, pggen.InsertOrganizationParams{
		ID:              pgtype.Text{String: org.ID(), Status: pgtype.Present},
		CreatedAt:       org.CreatedAt(),
		UpdatedAt:       org.UpdatedAt(),
		Name:            pgtype.Text{String: org.Name(), Status: pgtype.Present},
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
	})
	if err != nil {
		return databaseError(err)
	}
	return nil
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
	q := pggen.NewQuerier(tx)
	result, err := q.FindOrganizationByNameForUpdate(ctx, pgtype.Text{String: name, Status: pgtype.Present})
	if err != nil {
		return nil, err
	}
	org, err := otf.UnmarshalOrganizationDBResult(pggen.Organizations(result))
	if err != nil {
		return nil, err
	}
	if err := fn(org); err != nil {
		return nil, err
	}
	_, err = q.UpdateOrganizationByName(ctx, pggen.UpdateOrganizationByNameParams{
		Name:            pgtype.Text{String: name, Status: pgtype.Present},
		NewName:         pgtype.Text{String: org.Name(), Status: pgtype.Present},
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
		UpdatedAt:       org.UpdatedAt(),
	})
	if err != nil {
		return nil, err
	}
	return org, tx.Commit(ctx)
}

func (db OrganizationDB) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	q := pggen.NewQuerier(db.Pool)
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

	var items []*otf.Organization
	for _, r := range rows {
		org, err := otf.UnmarshalOrganizationDBResult(pggen.Organizations(r))
		if err != nil {
			return nil, err
		}
		items = append(items, org)
	}

	return &otf.OrganizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db OrganizationDB) Get(name string) (*otf.Organization, error) {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()
	r, err := q.FindOrganizationByName(ctx, pgtype.Text{String: name, Status: pgtype.Present})
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalOrganizationDBResult(pggen.Organizations(r))
}

func (db OrganizationDB) Delete(name string) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

	result, err := q.DeleteOrganization(ctx, pgtype.Text{String: name, Status: pgtype.Present})
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
