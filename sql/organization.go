package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateOrganization persists an Organization to the DB.
func (db *DB) CreateOrganization(ctx context.Context, org *otf.Organization) error {
	q := pggen.NewQuerier(db)
	_, err := q.InsertOrganization(ctx, pggen.InsertOrganizationParams{
		ID:              String(org.ID()),
		CreatedAt:       Timestamptz(org.CreatedAt()),
		UpdatedAt:       Timestamptz(org.UpdatedAt()),
		Name:            String(org.Name()),
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
	})
	if err != nil {
		return databaseError(err)
	}
	return nil
}

// UpdateOrganization persists an updated Organization to the DB. The existing
// org is fetched from the DB, the supplied func is invoked on the org, and the
// updated org is persisted back to the DB.
func (db *DB) UpdateOrganization(ctx context.Context, name string, fn func(*otf.Organization) error) (*otf.Organization, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	q := pggen.NewQuerier(tx)
	result, err := q.FindOrganizationByNameForUpdate(ctx, String(name))
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
		Name:            String(name),
		NewName:         String(org.Name()),
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
		UpdatedAt:       Timestamptz(org.UpdatedAt()),
	})
	if err != nil {
		return nil, err
	}
	return org, tx.Commit(ctx)
}

func (db *DB) ListOrganizations(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
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

func (db *DB) GetOrganization(ctx context.Context, name string) (*otf.Organization, error) {
	r, err := db.FindOrganizationByName(ctx, String(name))
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalOrganizationDBResult(pggen.Organizations(r))
}

func (db *DB) DeleteOrganization(ctx context.Context, name string) error {
	_, err := db.Querier.DeleteOrganization(ctx, String(name))
	if err != nil {
		return databaseError(err)
	}
	return nil
}
