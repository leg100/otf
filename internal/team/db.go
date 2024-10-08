package team

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

// TeamRow represents the result of a database query for a team.
type TeamRow struct {
	TeamID                          pgtype.Text
	Name                            pgtype.Text
	CreatedAt                       pgtype.Timestamptz
	PermissionManageWorkspaces      pgtype.Bool
	PermissionManageVCS             pgtype.Bool
	PermissionManageModules         pgtype.Bool
	OrganizationName                pgtype.Text
	SSOTeamID                       pgtype.Text
	Visibility                      pgtype.Text
	PermissionManagePolicies        pgtype.Bool
	PermissionManagePolicyOverrides pgtype.Bool
	PermissionManageProviders       pgtype.Bool
}

func (row TeamRow) ToTeam() *Team {
	to := Team{
		ID:           row.TeamID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		Name:         row.Name.String,
		Organization: row.OrganizationName.String,
		Visibility:   row.Visibility.String,
		Access: OrganizationAccess{
			ManageWorkspaces:      row.PermissionManageWorkspaces.Bool,
			ManageVCS:             row.PermissionManageVCS.Bool,
			ManageModules:         row.PermissionManageModules.Bool,
			ManageProviders:       row.PermissionManageProviders.Bool,
			ManagePolicies:        row.PermissionManagePolicies.Bool,
			ManagePolicyOverrides: row.PermissionManagePolicyOverrides.Bool,
		},
	}
	if row.SSOTeamID.Valid {
		to.SSOTeamID = &row.SSOTeamID.String
	}
	return &to
}

// pgdb stores team resources in a postgres database
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
	logr.Logger
}

func (db *pgdb) createTeam(ctx context.Context, team *Team) error {
	err := db.Querier(ctx).InsertTeam(ctx, sqlc.InsertTeamParams{
		ID:                              sql.String(team.ID),
		Name:                            sql.String(team.Name),
		CreatedAt:                       sql.Timestamptz(team.CreatedAt),
		OrganizationName:                sql.String(team.Organization),
		Visibility:                      sql.String(team.Visibility),
		SSOTeamID:                       sql.StringPtr(team.SSOTeamID),
		PermissionManageWorkspaces:      sql.Bool(team.Access.ManageWorkspaces),
		PermissionManageVCS:             sql.Bool(team.Access.ManageVCS),
		PermissionManageModules:         sql.Bool(team.Access.ManageModules),
		PermissionManageProviders:       sql.Bool(team.Access.ManageProviders),
		PermissionManagePolicies:        sql.Bool(team.Access.ManagePolicies),
		PermissionManagePolicyOverrides: sql.Bool(team.Access.ManagePolicyOverrides),
	})
	return sql.Error(err)
}

func (db *pgdb) UpdateTeam(ctx context.Context, teamID string, fn func(*Team) error) (*Team, error) {
	var team *Team
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		var err error

		// retrieve team
		result, err := q.FindTeamByIDForUpdate(ctx, sql.String(teamID))
		if err != nil {
			return err
		}
		team = TeamRow(result).ToTeam()

		// update team
		if err := fn(team); err != nil {
			return err
		}
		// persist update
		_, err = q.UpdateTeamByID(ctx, sqlc.UpdateTeamByIDParams{
			TeamID:                          sql.String(teamID),
			Name:                            sql.String(team.Name),
			Visibility:                      sql.String(team.Visibility),
			SSOTeamID:                       sql.StringPtr(team.SSOTeamID),
			PermissionManageWorkspaces:      sql.Bool(team.Access.ManageWorkspaces),
			PermissionManageVCS:             sql.Bool(team.Access.ManageVCS),
			PermissionManageModules:         sql.Bool(team.Access.ManageModules),
			PermissionManageProviders:       sql.Bool(team.Access.ManageProviders),
			PermissionManagePolicies:        sql.Bool(team.Access.ManagePolicies),
			PermissionManagePolicyOverrides: sql.Bool(team.Access.ManagePolicyOverrides),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return team, err
}

func (db *pgdb) getTeam(ctx context.Context, name, organization string) (*Team, error) {
	result, err := db.Querier(ctx).FindTeamByName(ctx, sqlc.FindTeamByNameParams{
		Name:             sql.String(name),
		OrganizationName: sql.String(organization),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return TeamRow(result).ToTeam(), nil
}

func (db *pgdb) getTeamByID(ctx context.Context, id string) (*Team, error) {
	result, err := db.Querier(ctx).FindTeamByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return TeamRow(result).ToTeam(), nil
}

func (db *pgdb) getTeamByTokenID(ctx context.Context, tokenID string) (*Team, error) {
	result, err := db.Querier(ctx).FindTeamByTokenID(ctx, sql.String(tokenID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return TeamRow(result).ToTeam(), nil
}

func (db *pgdb) listTeams(ctx context.Context, organization string) ([]*Team, error) {
	result, err := db.Querier(ctx).FindTeamsByOrg(ctx, sql.String(organization))
	if err != nil {
		return nil, err
	}

	items := make([]*Team, len(result))
	for i, r := range result {
		items[i] = TeamRow(r).ToTeam()
	}
	return items, nil
}

func (db *pgdb) deleteTeam(ctx context.Context, teamID string) error {
	_, err := db.Querier(ctx).DeleteTeamByID(ctx, sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

//
// Team tokens
//

func (db *pgdb) createTeamToken(ctx context.Context, token *Token) error {
	err := db.Querier(ctx).InsertTeamToken(ctx, sqlc.InsertTeamTokenParams{
		TeamTokenID: sql.String(token.ID),
		TeamID:      sql.String(token.TeamID),
		CreatedAt:   sql.Timestamptz(token.CreatedAt),
		Expiry:      sql.TimestamptzPtr(token.Expiry),
	})
	return err
}

func (db *pgdb) getTeamTokenByTeamID(ctx context.Context, teamID string) (*Token, error) {
	// query only returns 0 or 1 tokens
	result, err := db.Querier(ctx).FindTeamTokensByID(ctx, sql.String(teamID))
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	ot := &Token{
		ID:        result[0].TeamTokenID.String,
		CreatedAt: result[0].CreatedAt.Time.UTC(),
		TeamID:    result[0].TeamID.String,
	}
	if result[0].Expiry.Valid {
		ot.Expiry = internal.Time(result[0].Expiry.Time.UTC())
	}
	return ot, nil
}

func (db *pgdb) deleteTeamToken(ctx context.Context, team string) error {
	_, err := db.Querier(ctx).DeleteTeamTokenByID(ctx, sql.String(team))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
