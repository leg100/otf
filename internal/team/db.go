package team

import (
	"context"
	"errors"
	"time"

	"github.com/leg100/otf/internal/logr"
	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// pgdb stores team resources in a postgres database
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
	logr.Logger
}

func (db *pgdb) createTeam(ctx context.Context, team *Team) error {
	_, err := db.Exec(ctx, `
INSERT INTO teams (
    team_id,
    name,
    created_at,
    organization_name,
    visibility,
    sso_team_id,
    permission_manage_workspaces,
    permission_manage_vcs,
    permission_manage_modules,
    permission_manage_providers,
    permission_manage_policies,
    permission_manage_policy_overrides
) VALUES (
    @id,
    @name,
    @created_at,
    @organization,
    @visibility,
    @sso_team_id,
    @permission_manage_workspaces,
    @permission_manage_vcs,
    @permission_manage_modules,
    @permission_manage_providers,
    @permission_manage_policies,
    @permission_manage_policy_overrides
)
`,
		pgx.NamedArgs{
			"id":                                 team.ID,
			"name":                               team.Name,
			"created_at":                         team.CreatedAt,
			"organization":                       team.Organization,
			"visibility":                         team.Visibility,
			"sso_team_id":                        team.SSOTeamID,
			"permission_manage_workspaces":       team.ManageWorkspaces,
			"permission_manage_vcs":              team.ManageVCS,
			"permission_manage_modules":          team.ManageModules,
			"permission_manage_providers":        team.ManageProviders,
			"permission_manage_policies":         team.ManagePolicies,
			"permission_manage_policy_overrides": team.ManagePolicyOverrides,
		},
	)
	return err
}

func (db *pgdb) UpdateTeam(ctx context.Context, teamID resource.ID, fn func(context.Context, *Team) error) (*Team, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context) (*Team, error) {
			rows := db.Query(ctx, `
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams t
WHERE team_id = $1
FOR UPDATE OF t
`, teamID)
			return sql.CollectOneRow(rows, scan)
		},
		fn,
		func(ctx context.Context, team *Team) error {
			_, err := db.Exec(ctx, `
UPDATE teams
SET
    name = @name,
    visibility = @visibility,
    sso_team_id = @sso_team_id,
    permission_manage_workspaces = @permission_manage_workspaces,
    permission_manage_vcs = @permission_manage_vcs,
    permission_manage_modules = @permission_manage_modules,
    permission_manage_providers = @permission_manage_providers,
    permission_manage_policies = @permission_manage_policies,
    permission_manage_policy_overrides = @permission_manage_policy_overrides
WHERE team_id = @id
RETURNING team_id
`,
				pgx.NamedArgs{
					"id":                                 team.ID,
					"name":                               team.Name,
					"visibility":                         team.Visibility,
					"sso_team_id":                        team.SSOTeamID,
					"permission_manage_workspaces":       team.ManageWorkspaces,
					"permission_manage_vcs":              team.ManageVCS,
					"permission_manage_modules":          team.ManageModules,
					"permission_manage_providers":        team.ManageProviders,
					"permission_manage_policies":         team.ManagePolicies,
					"permission_manage_policy_overrides": team.ManagePolicyOverrides,
				},
			)
			return err
		},
	)
}

func (db *pgdb) getTeam(ctx context.Context, name string, organization organization.Name) (*Team, error) {
	rows := db.Query(ctx, `
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams
WHERE name              = $1
AND   organization_name = $2
`, name, organization)
	return sql.CollectOneRow(rows, scan)
}

func (db *pgdb) getTeamByID(ctx context.Context, id resource.ID) (*Team, error) {
	rows := db.Query(ctx, `
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams
WHERE team_id = $1
`, id)
	return sql.CollectOneRow(rows, scan)
}

func (db *pgdb) getTeamByTokenID(ctx context.Context, tokenID resource.TfeID) (*Team, error) {
	rows := db.Query(ctx, `
SELECT t.team_id, t.name, t.created_at, t.permission_manage_workspaces, t.permission_manage_vcs, t.permission_manage_modules, t.organization_name, t.sso_team_id, t.visibility, t.permission_manage_policies, t.permission_manage_policy_overrides, t.permission_manage_providers
FROM teams t
JOIN team_tokens tt USING (team_id)
WHERE tt.team_token_id = $1
`, tokenID)
	return sql.CollectOneRow(rows, scan)
}

func (db *pgdb) listTeams(ctx context.Context, organization organization.Name) ([]*Team, error) {
	rows := db.Query(ctx, `
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams
WHERE organization_name = $1
`, organization)
	return sql.CollectRows(rows, scan)
}

func (db *pgdb) deleteTeam(ctx context.Context, teamID resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM teams
WHERE team_id = $1
RETURNING team_id
`, teamID)
	return err
}

//
// Team tokens
//

func (db *pgdb) createTeamToken(ctx context.Context, token *Token) error {
	_, err := db.Exec(ctx, `
INSERT INTO team_tokens (
    team_token_id,
    created_at,
    team_id,
    expiry
) VALUES (
    @team_token_id,
    @created_at,
    @team_id,
    @expiry
) ON CONFLICT (team_id) DO UPDATE
  SET team_token_id = @team_token_id,
      created_at    = @created_at,
      expiry        = @expiry
`,
		pgx.NamedArgs{
			"team_token_id": token.ID,
			"team_id":       token.TeamID,
			"created_at":    token.CreatedAt,
			"expiry":        token.Expiry,
		})
	return err
}

func (db *pgdb) getTeamTokenByTeamID(ctx context.Context, teamID resource.TfeID) (*Token, error) {
	// query only returns 0 or 1 tokens
	rows := db.Query(ctx, `
SELECT *
FROM team_tokens
WHERE team_id = $1
`, teamID)
	token, err := pgx.CollectExactlyOneRow(rows, scanToken)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, internal.ErrResourceNotFound
	} else if err != nil {
		return nil, err
	}
	return token, nil
}

func (db *pgdb) deleteTeamToken(ctx context.Context, teamID resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM team_tokens
WHERE team_id = $1
`, teamID)
	return err
}

// Order of fields must match order of columns
type Model struct {
	ID                    resource.TfeID    `db:"team_id"`
	Name                  string            `db:"name"`
	CreatedAt             time.Time         `db:"created_at"`
	ManageWorkspaces      bool              `db:"permission_manage_workspaces"`
	ManageVCS             bool              `db:"permission_manage_vcs"`
	ManageModules         bool              `db:"permission_manage_modules"`
	Organization          organization.Name `db:"organization_name"`
	SSOTeamID             *string           `db:"sso_team_id"`
	Visibility            string
	ManagePolicies        bool `db:"permission_manage_policies"`
	ManagePolicyOverrides bool `db:"permission_manage_policy_overrides"`
	ManageProviders       bool `db:"permission_manage_providers"`
}

func (m Model) ToTeam() *Team {
	return &Team{
		ID:                    m.ID,
		Name:                  m.Name,
		CreatedAt:             m.CreatedAt,
		ManageWorkspaces:      m.ManageWorkspaces,
		ManageModules:         m.ManageModules,
		ManageVCS:             m.ManageVCS,
		Organization:          m.Organization,
		SSOTeamID:             m.SSOTeamID,
		ManagePolicies:        m.ManagePolicies,
		ManagePolicyOverrides: m.ManagePolicyOverrides,
		ManageProviders:       m.ManageProviders,
		Visibility:            m.Visibility,
	}
}

func scan(row pgx.CollectableRow) (*Team, error) {
	m, err := pgx.RowToAddrOfStructByName[Model](row)
	if err != nil {
		return nil, err
	}
	return m.ToTeam(), nil
}

func scanToken(row pgx.CollectableRow) (*Token, error) {
	return pgx.RowToAddrOfStructByName[Token](row)
}
