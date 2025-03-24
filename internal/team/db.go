package team

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
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
    @permission_manage_policy_overrides,
)
`,
		pgx.NamedArgs{
			"id":                                 team.ID,
			"name":                               team.Name,
			"created_at":                         team.CreatedAt,
			"organization":                       team.Organization,
			"visibility":                         team.Visibility,
			"sso_team_id":                        team.SSOTeamID,
			"permission_manage_workspaces":       team.Access.ManageWorkspaces,
			"permission_manage_vcs":              team.Access.ManageWorkspaces,
			"permission_manage_modules":          team.Access.ManageWorkspaces,
			"permission_manage_providers":        team.Access.ManageWorkspaces,
			"permission_manage_policies":         team.Access.ManageWorkspaces,
			"permission_manage_policy_overrides": team.Access.ManageWorkspaces,
		},
	)
	return err
}

func (db *pgdb) UpdateTeam(ctx context.Context, teamID resource.TfeID, fn func(context.Context, *Team) error) (*Team, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Team, error) {
			rows := db.Query(ctx, `
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams t
WHERE team_id = $1
FOR UPDATE OF t
`, teamID)
			return sql.CollectOneRow(rows, db.scan)
		},
		fn,
		func(ctx context.Context, conn sql.Connection, team *Team) error {
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
					"permission_manage_workspaces":       team.Access.ManageWorkspaces,
					"permission_manage_vcs":              team.Access.ManageWorkspaces,
					"permission_manage_modules":          team.Access.ManageWorkspaces,
					"permission_manage_providers":        team.Access.ManageWorkspaces,
					"permission_manage_policies":         team.Access.ManageWorkspaces,
					"permission_manage_policy_overrides": team.Access.ManageWorkspaces,
				},
			)
			return err
		},
	)
}

func (db *pgdb) getTeam(ctx context.Context, name string, organization resource.OrganizationName) (*Team, error) {
	rows := db.Query(ctx, `
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams
WHERE name              = $1
AND   organization_name = $2
`, name, organization)
	return sql.CollectOneRow(rows, db.scan)
}

func (db *pgdb) getTeamByID(ctx context.Context, id resource.TfeID) (*Team, error) {
	rows := db.Query(ctx, `
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams
WHERE team_id = $1
`, id)
	return sql.CollectOneRow(rows, db.scan)
}

func (db *pgdb) getTeamByTokenID(ctx context.Context, tokenID resource.TfeID) (*Team, error) {
	rows := db.Query(ctx, `
SELECT t.team_id, t.name, t.created_at, t.permission_manage_workspaces, t.permission_manage_vcs, t.permission_manage_modules, t.organization_name, t.sso_team_id, t.visibility, t.permission_manage_policies, t.permission_manage_policy_overrides, t.permission_manage_providers
FROM teams t
JOIN team_tokens tt USING (team_id)
WHERE tt.team_token_id = $1
`, tokenID)
	return sql.CollectOneRow(rows, db.scan)
}

func (db *pgdb) listTeams(ctx context.Context, organization resource.OrganizationName) ([]*Team, error) {
	rows := db.Query(ctx, `
SELECT team_id, name, created_at, permission_manage_workspaces, permission_manage_vcs, permission_manage_modules, organization_name, sso_team_id, visibility, permission_manage_policies, permission_manage_policy_overrides, permission_manage_providers
FROM teams
WHERE organization_name = $1
`, organization)
	return sql.CollectRows(rows, db.scan)
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
    $created_at,
    $team_id,
    $expiry
) ON CONFLICT (team_id) DO UPDATE
  SET team_token_id = @team_token_id,
      created_at    = @created_at,
      expiry        = @expiry
`,
		pgx.NamedArgs{
			"team_token_id": token.TfeID,
			"team_id":       token.TeamID,
			"created_at":    token.CreatedAt,
			"expiry":        token.Expiry,
		})
	return err
}

func (db *pgdb) getTeamTokenByTeamID(ctx context.Context, teamID resource.TfeID) (*Token, error) {
	// query only returns 0 or 1 tokens
	rows := db.Query(ctx, `
SELECT team_token_id, description, created_at, team_id, expiry
FROM team_tokens
WHERE team_id = $1
`, teamID)
	token, err := pgx.CollectExactlyOneRow(rows, db.scanToken)
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

func (db *pgdb) scan(row pgx.CollectableRow) (*Team, error) {
	team, err := pgx.RowToAddrOfStructByName[Team](row)
	if err != nil {
		return nil, err
	}
	team.CreatedAt = team.CreatedAt.UTC()
	return team, nil
}

func (db *pgdb) scanToken(row pgx.CollectableRow) (*Token, error) {
	token, err := pgx.RowToAddrOfStructByName[Token](row)
	if err != nil {
		return nil, err
	}
	token.CreatedAt = token.CreatedAt.UTC()
	return token, nil
}
