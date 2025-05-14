package organization

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a database of organizations on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) create(ctx context.Context, org *Organization) error {
	_, err := db.Exec(ctx, `
		INSERT INTO organizations (
			organization_id,
			created_at,
			updated_at,
			name,
			email,
			collaborator_auth_policy,
			cost_estimation_enabled,
			session_remember,
			session_timeout,
			allow_force_delete_workspaces
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		org.ID,
		org.CreatedAt,
		org.UpdatedAt,
		org.Name,
		org.Email,
		org.CollaboratorAuthPolicy,
		org.CostEstimationEnabled,
		org.SessionRemember,
		org.SessionTimeout,
		org.AllowForceDeleteWorkspaces,
	)
	return err
}

func (db *pgdb) update(ctx context.Context, name Name, fn func(context.Context, *Organization) error) (*Organization, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context) (*Organization, error) {
			row := db.Query(ctx, `
SELECT *
FROM organizations
WHERE name = $1
FOR UPDATE`,
				name)
			return sql.CollectOneRow(row, db.scan)
		},
		fn,
		func(ctx context.Context, org *Organization) error {
			_, err := db.Exec(ctx, `
UPDATE organizations
SET
	name = $1,
	email = $2,
	collaborator_auth_policy = $3,
	cost_estimation_enabled = $4,
	session_remember = $5,
	session_timeout = $6,
	allow_force_delete_workspaces = $7,
	updated_at = $8
WHERE name = $9`,
				org.Name,
				org.Email,
				org.CollaboratorAuthPolicy,
				org.CostEstimationEnabled,
				org.SessionRemember,
				org.SessionTimeout,
				org.AllowForceDeleteWorkspaces,
				org.UpdatedAt,
				name,
			)
			return err
		},
	)
}

// dbListOptions represents the options for listing organizations via the
// database.
type dbListOptions struct {
	names []Name // filter organizations by name if non-nil
	resource.PageOptions
}

func (db *pgdb) list(ctx context.Context, opts dbListOptions) (*resource.Page[*Organization], error) {
	// Convert organization name type slice to string slice. By default, return
	// all organizations by specifying '%'.
	names := []string{"%"}
	if opts.names != nil {
		names = make([]string, len(opts.names))
		for i, name := range opts.names {
			names[i] = name.String()
		}
	}

	rows := db.Query(ctx, `
SELECT *
FROM organizations
WHERE name LIKE ANY($1::text[])
ORDER BY updated_at DESC
LIMIT $2::int OFFSET $3::int
`,
		sql.StringArray(names),
		sql.GetLimit(opts.PageOptions),
		sql.GetOffset(opts.PageOptions),
	)
	items, err := sql.CollectRows(rows, db.scan)
	if err != nil {
		return nil, err
	}

	count, err := db.Int(ctx, `
SELECT count(*)
FROM organizations
WHERE name LIKE ANY($1::text[])
`,
		sql.StringArray(names),
	)
	if err != nil {
		return nil, err
	}
	return resource.NewPage(items, opts.PageOptions, &count), nil
}

func (db *pgdb) get(ctx context.Context, name Name) (*Organization, error) {
	row := db.Query(ctx, ` SELECT * FROM organizations WHERE name = $1 `, name)
	return sql.CollectOneRow(row, db.scan)
}

func (db *pgdb) getByID(ctx context.Context, id resource.TfeID) (*Organization, error) {
	row := db.Query(ctx, ` SELECT * FROM organizations WHERE organization_id = $1 `, id)
	return sql.CollectOneRow(row, db.scan)
}

func (db *pgdb) delete(ctx context.Context, name Name) error {
	_, err := db.Exec(ctx, ` DELETE FROM organizations WHERE name = $1 `, name)
	return err
}

//
// Organization tokens
//

func (db *pgdb) upsertOrganizationToken(ctx context.Context, token *OrganizationToken) error {
	_, err := db.Exec(ctx, `
INSERT INTO organization_tokens (
    organization_token_id,
    created_at,
    organization_name,
    expiry
) VALUES (
    $1,
    $2,
    $3,
    $4
) ON CONFLICT (organization_name) DO UPDATE
  SET created_at            = $2,
      organization_token_id = $1,
      expiry                = $4`,
		token.ID,
		token.CreatedAt,
		token.Organization,
		token.Expiry,
	)
	return err
}

func (db *pgdb) getOrganizationTokenByName(ctx context.Context, organization Name) (*OrganizationToken, error) {
	row := db.Query(ctx, ` SELECT * FROM organization_tokens WHERE organization_name = $1 `, organization)
	return sql.CollectOneRow(row, db.scanToken)
}

func (db *pgdb) listOrganizationTokens(ctx context.Context, organization Name) ([]*OrganizationToken, error) {
	rows := db.Query(ctx, `
SELECT organization_token_id, created_at, organization_name, expiry
FROM organization_tokens
WHERE organization_name = $1
`,
		organization,
	)
	return sql.CollectRows(rows, db.scanToken)
}

func (db *pgdb) getOrganizationTokenByID(ctx context.Context, tokenID resource.TfeID) (*OrganizationToken, error) {
	row := db.Query(ctx, `
SELECT *
FROM organization_tokens
WHERE organization_token_id = $1
`,
		tokenID)
	return sql.CollectOneRow(row, db.scanToken)
}

func (db *pgdb) deleteOrganizationToken(ctx context.Context, organization Name) error {
	_, err := db.Exec(ctx, `
DELETE
FROM organization_tokens
WHERE organization_name = $1
`,
		organization)
	return err
}

func (db *pgdb) scan(row pgx.CollectableRow) (*Organization, error) {
	return pgx.RowToAddrOfStructByName[Organization](row)
}

func (db *pgdb) scanToken(row pgx.CollectableRow) (*OrganizationToken, error) {
	return pgx.RowToAddrOfStructByName[OrganizationToken](row)
}
