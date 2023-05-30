package organization

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is a database of organizations on postgres
	pgdb struct {
		internal.DB // provides access to generated SQL queries
	}

	row struct {
		OrganizationID  pgtype.Text        `json:"organization_id"`
		CreatedAt       pgtype.Timestamptz `json:"created_at"`
		UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
		Name            pgtype.Text        `json:"name"`
		SessionRemember int                `json:"session_remember"`
		SessionTimeout  int                `json:"session_timeout"`
	}
)

// GetByID implements pubsub.Getter
func (db *pgdb) GetByID(ctx context.Context, id string, action pubsub.DBAction) (any, error) {
	if action == pubsub.DeleteDBAction {
		return &Organization{ID: id}, nil
	}
	r, err := db.FindOrganizationByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return row(r).toOrganization(), nil
}

func (db *pgdb) update(ctx context.Context, name string, fn func(*Organization) error) (*Organization, error) {
	var org *Organization
	err := db.Tx(ctx, func(tx internal.DB) error {
		result, err := tx.FindOrganizationByNameForUpdate(ctx, sql.String(name))
		if err != nil {
			return err
		}
		org = row(result).toOrganization()

		if err := fn(org); err != nil {
			return err
		}
		_, err = tx.UpdateOrganizationByName(ctx, pggen.UpdateOrganizationByNameParams{
			Name:            sql.String(name),
			NewName:         sql.String(org.Name),
			SessionRemember: org.SessionRemember,
			SessionTimeout:  org.SessionTimeout,
			UpdatedAt:       sql.Timestamptz(org.UpdatedAt),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return org, err
}

func (db *pgdb) list(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error) {
	// optionally filter by organization name
	var names []string
	if opts.Names != nil {
		names = opts.Names
	} else {
		names = []string{"%"} // return all organizations
	}

	batch := &pgx.Batch{}

	db.FindOrganizationsBatch(batch, pggen.FindOrganizationsParams{
		Names:  names,
		Limit:  opts.GetLimit(),
		Offset: opts.GetOffset(),
	})
	db.CountOrganizationsBatch(batch, names)
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
		items = append(items, row(r).toOrganization())
	}

	return &OrganizationList{
		Items:      items,
		Pagination: internal.NewPagination(opts.ListOptions, count),
	}, nil
}

func (db *pgdb) get(ctx context.Context, name string) (*Organization, error) {
	r, err := db.FindOrganizationByName(ctx, sql.String(name))
	if err != nil {
		return nil, sql.Error(err)
	}
	return row(r).toOrganization(), nil
}

func (db *pgdb) delete(ctx context.Context, name string) error {
	_, err := db.DeleteOrganizationByName(ctx, sql.String(name))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// row converts an organization database row into an
// organization.
func (r row) toOrganization() *Organization {
	return &Organization{
		ID:              r.OrganizationID.String,
		CreatedAt:       r.CreatedAt.Time.UTC(),
		UpdatedAt:       r.UpdatedAt.Time.UTC(),
		Name:            r.Name.String,
		SessionRemember: r.SessionRemember,
		SessionTimeout:  r.SessionTimeout,
	}
}
