package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateOrganization persists an Organization to the DB.
func (db *DB) CreateOrganization(ctx context.Context, org *otf.Organization) error {
	_, err := db.InsertOrganization(ctx, pggen.InsertOrganizationParams{
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
	var org *otf.Organization
	err := db.tx(ctx, func(tx *DB) error {
		result, err := tx.FindOrganizationByNameForUpdate(ctx, String(name))
		if err != nil {
			return err
		}
		org = otf.UnmarshalOrganizationRow(pggen.Organizations(result))
		if err := fn(org); err != nil {
			return err
		}
		_, err = tx.UpdateOrganizationByName(ctx, pggen.UpdateOrganizationByNameParams{
			Name:            String(name),
			NewName:         String(org.Name()),
			SessionRemember: org.SessionRemember(),
			SessionTimeout:  org.SessionTimeout(),
			UpdatedAt:       Timestamptz(org.UpdatedAt()),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return org, err
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
		items = append(items, otf.UnmarshalOrganizationRow(pggen.Organizations(r)))
	}

	return &otf.OrganizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *DB) GetOrganizationNameByWorkspaceID(ctx context.Context, workspaceID string) (string, error) {
	name, err := db.FindOrganizationNameByWorkspaceID(ctx, String(workspaceID))
	if err != nil {
		return "", databaseError(err)
	}
	return name.String, nil
}

func (db *DB) GetOrganization(ctx context.Context, name string) (*otf.Organization, error) {
	r, err := db.FindOrganizationByName(ctx, String(name))
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalOrganizationRow(pggen.Organizations(r)), nil
}

func (db *DB) DeleteOrganization(ctx context.Context, name string) error {
	_, err := db.Querier.DeleteOrganization(ctx, String(name))
	if err != nil {
		return databaseError(err)
	}
	return nil
}
