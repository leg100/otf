package organization

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/workspace/execution"
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
			allow_force_delete_workspaces,
			default_execution_kind
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
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
		org.DefaultMode.Kind(),
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
			return sql.CollectOneRow(row, scan)
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
	default_execution_kind = $8,
	default_agent_pool_id = $9,
	updated_at = $10
WHERE name = $11`,
				org.Name,
				org.Email,
				org.CollaboratorAuthPolicy,
				org.CostEstimationEnabled,
				org.SessionRemember,
				org.SessionTimeout,
				org.AllowForceDeleteWorkspaces,
				org.DefaultMode.Kind(),
				org.DefaultMode.AgentPoolID(),
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
	items, err := sql.CollectRows(rows, scan)
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
	return sql.CollectOneRow(row, scan)
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

func (db *pgdb) scanToken(row pgx.CollectableRow) (*OrganizationToken, error) {
	return pgx.RowToAddrOfStructByName[OrganizationToken](row)
}

func scan(row pgx.CollectableRow) (*Organization, error) {
	type model struct {
		ID                         resource.TfeID  `db:"organization_id"`
		CreatedAt                  time.Time       `db:"created_at"`
		UpdatedAt                  time.Time       `db:"updated_at"`
		Name                       Name            `db:"name"`
		Email                      *string         `db:"email"`
		CollaboratorAuthPolicy     *string         `db:"collaborator_auth_policy"`
		SessionRemember            *int            `db:"session_remember"`
		SessionTimeout             *int            `db:"session_timeout"`
		AllowForceDeleteWorkspaces bool            `db:"allow_force_delete_workspaces"`
		CostEstimationEnabled      bool            `db:"cost_estimation_enabled"`
		DefaultAgentPoolID         *resource.TfeID `db:"default_agent_pool_id"`
		DefaultExecutionKind       execution.Kind  `db:"default_execution_kind"`
	}
	m, err := pgx.RowToStructByName[model](row)
	if err != nil {
		return nil, err
	}
	org := &Organization{
		ID:                         m.ID,
		CreatedAt:                  m.CreatedAt,
		UpdatedAt:                  m.UpdatedAt,
		Name:                       m.Name,
		Email:                      m.Email,
		CollaboratorAuthPolicy:     m.CollaboratorAuthPolicy,
		SessionRemember:            m.SessionRemember,
		SessionTimeout:             m.SessionTimeout,
		AllowForceDeleteWorkspaces: m.AllowForceDeleteWorkspaces,
		CostEstimationEnabled:      m.CostEstimationEnabled,
	}
	mode, err := execution.NewMode(m.DefaultExecutionKind, m.DefaultAgentPoolID)
	if err != nil {
		return nil, err
	}
	org.DefaultMode = mode

	return org, err
}
