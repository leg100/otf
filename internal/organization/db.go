package organization

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

type (
	// dbListOptions represents the options for listing organizations via the
	// database.
	dbListOptions struct {
		names []string // filter organizations by name if non-nil
		resource.PageOptions
	}
)

// row is the row result of a database query for organizations
type row struct {
	OrganizationID             resource.ID
	CreatedAt                  pgtype.Timestamptz
	UpdatedAt                  pgtype.Timestamptz
	Name                       Name
	SessionRemember            pgtype.Int4
	SessionTimeout             pgtype.Int4
	Email                      pgtype.Text
	CollaboratorAuthPolicy     pgtype.Text
	AllowForceDeleteWorkspaces pgtype.Bool
	CostEstimationEnabled      pgtype.Bool
}

// row converts an organization database row into an
// organization.
func (r row) toOrganization() *Organization {
	org := &Organization{
		ID:                         r.OrganizationID,
		CreatedAt:                  r.CreatedAt.Time.UTC(),
		UpdatedAt:                  r.UpdatedAt.Time.UTC(),
		Name:                       r.Name,
		AllowForceDeleteWorkspaces: r.AllowForceDeleteWorkspaces.Bool,
		CostEstimationEnabled:      r.CostEstimationEnabled.Bool,
	}
	if r.SessionRemember.Valid {
		sessionRememberInt := int(r.SessionRemember.Int32)
		org.SessionRemember = &sessionRememberInt
	}
	if r.SessionTimeout.Valid {
		sessionTimeoutInt := int(r.SessionTimeout.Int32)
		org.SessionTimeout = &sessionTimeoutInt
	}
	if r.Email.Valid {
		org.Email = &r.Email.String
	}
	if r.CollaboratorAuthPolicy.Valid {
		org.CollaboratorAuthPolicy = &r.CollaboratorAuthPolicy.String
	}
	return org
}

// pgdb is a database of organizations on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) create(ctx context.Context, org *Organization) error {
	err := q.InsertOrganization(ctx, db.Conn(ctx), InsertOrganizationParams{
		ID:                         org.ID,
		CreatedAt:                  sql.Timestamptz(org.CreatedAt),
		UpdatedAt:                  sql.Timestamptz(org.UpdatedAt),
		Name:                       org.Name,
		SessionRemember:            sql.Int4Ptr(org.SessionRemember),
		SessionTimeout:             sql.Int4Ptr(org.SessionTimeout),
		Email:                      sql.StringPtr(org.Email),
		CollaboratorAuthPolicy:     sql.StringPtr(org.CollaboratorAuthPolicy),
		CostEstimationEnabled:      sql.Bool(org.CostEstimationEnabled),
		AllowForceDeleteWorkspaces: sql.Bool(org.AllowForceDeleteWorkspaces),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) update(ctx context.Context, name Name, fn func(context.Context, *Organization) error) (*Organization, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Organization, error) {
			result, err := q.FindOrganizationByNameForUpdate(ctx, conn, name)
			if err != nil {
				return nil, err
			}
			return row(result).toOrganization(), nil
		},
		fn,
		func(ctx context.Context, conn sql.Connection, org *Organization) error {
			_, err := q.UpdateOrganizationByName(ctx, conn, UpdateOrganizationByNameParams{
				Name:                       name,
				NewName:                    org.Name,
				Email:                      sql.StringPtr(org.Email),
				CollaboratorAuthPolicy:     sql.StringPtr(org.CollaboratorAuthPolicy),
				CostEstimationEnabled:      sql.Bool(org.CostEstimationEnabled),
				SessionRemember:            sql.Int4Ptr(org.SessionRemember),
				SessionTimeout:             sql.Int4Ptr(org.SessionTimeout),
				UpdatedAt:                  sql.Timestamptz(org.UpdatedAt),
				AllowForceDeleteWorkspaces: sql.Bool(org.AllowForceDeleteWorkspaces),
			})
			return err
		},
	)
}

func (db *pgdb) list(ctx context.Context, opts dbListOptions) (*resource.Page[*Organization], error) {
	if opts.names == nil {
		opts.names = []string{"%"} // return all organizations
	}

	rows, err := q.FindOrganizations(ctx, db.Conn(ctx), FindOrganizationsParams{
		Names:  sql.StringArray(opts.names),
		Limit:  sql.GetLimit(opts.PageOptions),
		Offset: sql.GetOffset(opts.PageOptions),
	})
	if err != nil {
		return nil, err
	}
	count, err := q.CountOrganizations(ctx, db.Conn(ctx), sql.StringArray(opts.names))
	if err != nil {
		return nil, err
	}

	items := make([]*Organization, len(rows))
	for i, r := range rows {
		items[i] = row(r).toOrganization()
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}

func (db *pgdb) get(ctx context.Context, name string) (*Organization, error) {
	r, err := q.FindOrganizationByName(ctx, db.Conn(ctx), sql.String(name))
	if err != nil {
		return nil, sql.Error(err)
	}
	return row(r).toOrganization(), nil
}

func (db *pgdb) getByID(ctx context.Context, id resource.ID) (*Organization, error) {
	r, err := q.FindOrganizationByID(ctx, db.Conn(ctx), id)
	if err != nil {
		return nil, sql.Error(err)
	}
	return row(r).toOrganization(), nil
}

func (db *pgdb) delete(ctx context.Context, name string) error {
	_, err := q.DeleteOrganizationByName(ctx, db.Conn(ctx), sql.String(name))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

//
// Organization tokens
//

// tokenRow is the row result of a database query for organization tokens
type tokenRow struct {
	OrganizationTokenID resource.ID        `json:"organization_token_id"`
	CreatedAt           pgtype.Timestamptz `json:"created_at"`
	OrganizationName    pgtype.Text        `json:"organization_name"`
	Expiry              pgtype.Timestamptz `json:"expiry"`
}

func (result tokenRow) toToken() *OrganizationToken {
	ot := &OrganizationToken{
		ID:           result.OrganizationTokenID,
		CreatedAt:    result.CreatedAt.Time.UTC(),
		Organization: result.OrganizationName.String,
	}
	if result.Expiry.Valid {
		ot.Expiry = internal.Time(result.Expiry.Time.UTC())
	}
	return ot
}

func (db *pgdb) upsertOrganizationToken(ctx context.Context, token *OrganizationToken) error {
	err := q.UpsertOrganizationToken(ctx, db.Conn(ctx), UpsertOrganizationTokenParams{
		OrganizationTokenID: token.ID,
		OrganizationName:    token.Organization,
		CreatedAt:           sql.Timestamptz(token.CreatedAt),
		Expiry:              sql.TimestamptzPtr(token.Expiry),
	})
	return err
}

func (db *pgdb) getOrganizationTokenByName(ctx context.Context, organization Name) (*OrganizationToken, error) {
	result, err := q.FindOrganizationTokensByName(ctx, db.Conn(ctx), organization)
	if err != nil {
		return nil, sql.Error(err)
	}
	return tokenRow(result).toToken(), nil
}

func (db *pgdb) listOrganizationTokens(ctx context.Context, organization Name) ([]*OrganizationToken, error) {
	result, err := q.FindOrganizationTokens(ctx, db.Conn(ctx), organization)
	if err != nil {
		return nil, sql.Error(err)
	}
	items := make([]*OrganizationToken, len(result))
	for i, r := range result {
		items[i] = tokenRow(r).toToken()
	}
	return items, nil
}

func (db *pgdb) getOrganizationTokenByID(ctx context.Context, tokenID resource.ID) (*OrganizationToken, error) {
	result, err := q.FindOrganizationTokensByID(ctx, db.Conn(ctx), tokenID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return tokenRow(result).toToken(), nil
}

func (db *pgdb) deleteOrganizationToken(ctx context.Context, organization Name) error {
	_, err := q.DeleteOrganiationTokenByName(ctx, db.Conn(ctx), organization)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
