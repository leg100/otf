package organization

import (
	"context"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a database of organizations on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) create(ctx context.Context, org *Organization) error {
	_, err := db.Conn(ctx).Exec(ctx, `
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
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) update(ctx context.Context, name resource.OrganizationName, fn func(context.Context, *Organization) error) (*Organization, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Organization, error) {
			row := db.Conn(ctx).QueryRow(ctx, `
SELECT *
FROM organizations
WHERE name = $1
FOR UPDATE`,
				name)
			return db.scan(row)
		},
		fn,
		func(ctx context.Context, conn sql.Connection, org *Organization) error {
			_, err := db.Conn(ctx).Exec(ctx, `
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
	names []resource.OrganizationName // filter organizations by name if non-nil
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

	rows, err := db.Conn(ctx).Query(ctx, `
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
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Organization
	for rows.Next() {
		org, err := db.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, org)
	}
	countRow := db.Conn(ctx).QueryRow(ctx, `
SELECT count(*)
FROM organizations
WHERE name LIKE ANY($1::text[])
`,
		sql.StringArray(names),
	)
	var count int64
	if err := countRow.Scan(&count); err != nil {
		return nil, err
	}
	return resource.NewPage(items, opts.PageOptions, &count), nil
}

func (db *pgdb) get(ctx context.Context, name resource.OrganizationName) (*Organization, error) {
	row := db.Conn(ctx).QueryRow(ctx, `
SELECT * FROM organizations WHERE name = $1
`,
		name)
	return db.scan(row)
}

func (db *pgdb) getByID(ctx context.Context, id resource.TfeID) (*Organization, error) {
	row := db.Conn(ctx).QueryRow(ctx, `
SELECT * FROM organizations WHERE organization_id = $1
`,
		id)
	return db.scan(row)
}

func (db *pgdb) delete(ctx context.Context, name resource.OrganizationName) error {
	_, err := db.Conn(ctx).Exec(ctx, `
DELETE
FROM organizations
WHERE name = $1
`,
		name)
	return sql.Error(err)
}

//
// Organization tokens
//

func (db *pgdb) upsertOrganizationToken(ctx context.Context, token *OrganizationToken) error {
	_, err := db.Conn(ctx).Exec(ctx, `
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

func (db *pgdb) getOrganizationTokenByName(ctx context.Context, organization resource.OrganizationName) (*OrganizationToken, error) {
	row := db.Conn(ctx).QueryRow(ctx, `
SELECT *
FROM organization_tokens
WHERE organization_name = $1
`,
		organization)
	return db.scanToken(row)
}

func (db *pgdb) listOrganizationTokens(ctx context.Context, organization resource.OrganizationName) ([]*OrganizationToken, error) {
	rows, err := db.Conn(ctx).Query(ctx, `
SELECT organization_token_id, created_at, organization_name, expiry
FROM organization_tokens
WHERE organization_name = $1
`,
		organization,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*OrganizationToken
	for rows.Next() {
		org, err := db.scanToken(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, org)
	}
	return items, nil
}

func (db *pgdb) getOrganizationTokenByID(ctx context.Context, tokenID resource.TfeID) (*OrganizationToken, error) {
	row := db.Conn(ctx).QueryRow(ctx, `
SELECT *
FROM organization_tokens
WHERE organization_token_id = $1
`,
		tokenID)
	return db.scanToken(row)
}

func (db *pgdb) deleteOrganizationToken(ctx context.Context, organization resource.OrganizationName) error {
	_, err := db.Conn(ctx).Exec(ctx, `
DELETE
FROM organization_tokens
WHERE organization_name = $1
`,
		organization)
	return sql.Error(err)
}

func (db *pgdb) scan(row sql.Scanner) (*Organization, error) {
	var org Organization
	err := row.Scan(
		&org.ID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.Name,
		&org.SessionRemember,
		&org.SessionTimeout,
		&org.Email,
		&org.CollaboratorAuthPolicy,
		&org.AllowForceDeleteWorkspaces,
		&org.CostEstimationEnabled,
	)
	org.CreatedAt = org.CreatedAt.UTC()
	org.UpdatedAt = org.UpdatedAt.UTC()
	return &org, sql.Error(err)
}

func (db *pgdb) scanToken(row sql.Scanner) (*OrganizationToken, error) {
	var token OrganizationToken
	err := row.Scan(
		&token.ID,
		&token.CreatedAt,
		&token.Organization,
		&token.Expiry,
	)
	token.CreatedAt = token.CreatedAt.UTC()
	return &token, sql.Error(err)
}

//type scanResource[T any] interface {
//	scan(scanner) (T, error)
//}
//
//func scanRows[T any](ctx context.Context, conn sql.Connection, scanner scanResource[T], query string, params ...any) ([]T, error) {
//	rows, err := conn.Query(ctx, query, params...)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	var items []T
//	for rows.Next() {
//		item, err := scanner.scan(rows)
//		if err != nil {
//			return nil, err
//		}
//		items = append(items, item)
//	}
//	return items, nil
//}
