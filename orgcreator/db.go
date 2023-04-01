package orgcreator

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is a database of organizations on postgres
type pgdb struct {
	otf.DB // provides access to generated SQL queries
}

func (db *pgdb) create(ctx context.Context, org *organization.Organization) error {
	_, err := db.InsertOrganization(ctx, pggen.InsertOrganizationParams{
		ID:              sql.String(org.ID),
		CreatedAt:       sql.Timestamptz(org.CreatedAt),
		UpdatedAt:       sql.Timestamptz(org.UpdatedAt),
		Name:            sql.String(org.Name),
		SessionRemember: org.SessionRemember,
		SessionTimeout:  org.SessionTimeout,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
