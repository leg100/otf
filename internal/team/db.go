package team

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

// TeamRow represents the result of a database query for a team.
type TeamRow struct {
	TeamID                          resource.TfeID
	Name                            pgtype.Text
	CreatedAt                       pgtype.Timestamptz
	PermissionManageWorkspaces      pgtype.Bool
	PermissionManageVCS             pgtype.Bool
	PermissionManageModules         pgtype.Bool
	OrganizationName                resource.OrganizationName
	SSOTeamID                       pgtype.Text
	Visibility                      pgtype.Text
	PermissionManagePolicies        pgtype.Bool
	PermissionManagePolicyOverrides pgtype.Bool
	PermissionManageProviders       pgtype.Bool
}

func (row TeamRow) ToTeam() *Team {
	to := Team{
		ID:           row.TeamID,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		Name:         row.Name.String,
		Organization: row.OrganizationName,
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
	err := q.InsertTeam(ctx, db.Conn(ctx), InsertTeamParams{
		ID:                              team.ID,
		Name:                            sql.String(team.Name),
		CreatedAt:                       sql.Timestamptz(team.CreatedAt),
		OrganizationName:                team.Organization,
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

func (db *pgdb) UpdateTeam(ctx context.Context, teamID resource.ID, fn func(context.Context, *Team) error) (*Team, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Team, error) {
			result, err := q.FindTeamByIDForUpdate(ctx, db.Conn(ctx), teamID)
			if err != nil {
				return nil, err
			}
			return TeamRow(result).ToTeam(), nil
		},
		fn,
		func(ctx context.Context, conn sql.Connection, team *Team) error {
			_, err := q.UpdateTeamByID(ctx, conn, UpdateTeamByIDParams{
				TeamID:                          teamID,
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
			return err
		},
	)
}

func (db *pgdb) getTeam(ctx context.Context, name string, organization resource.OrganizationName) (*Team, error) {
	result, err := q.FindTeamByName(ctx, db.Conn(ctx), FindTeamByNameParams{
		Name:             sql.String(name),
		OrganizationName: organization,
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return TeamRow(result).ToTeam(), nil
}

func (db *pgdb) getTeamByID(ctx context.Context, id resource.ID) (*Team, error) {
	result, err := q.FindTeamByID(ctx, db.Conn(ctx), id.(*resource.TfeID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return TeamRow(result).ToTeam(), nil
}

func (db *pgdb) getTeamByTokenID(ctx context.Context, tokenID resource.TfeID) (*Team, error) {
	result, err := q.FindTeamByTokenID(ctx, db.Conn(ctx), tokenID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return TeamRow(result).ToTeam(), nil
}

func (db *pgdb) listTeams(ctx context.Context, organization resource.OrganizationName) ([]*Team, error) {
	result, err := q.FindTeamsByOrg(ctx, db.Conn(ctx), organization)
	if err != nil {
		return nil, err
	}

	items := make([]*Team, len(result))
	for i, r := range result {
		items[i] = TeamRow(r).ToTeam()
	}
	return items, nil
}

func (db *pgdb) deleteTeam(ctx context.Context, teamID resource.TfeID) error {
	_, err := q.DeleteTeamByID(ctx, db.Conn(ctx), teamID)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

//
// Team tokens
//

func (db *pgdb) createTeamToken(ctx context.Context, token *Token) error {
	err := q.InsertTeamToken(ctx, db.Conn(ctx), InsertTeamTokenParams{
		TeamTokenID: token.TfeID,
		TeamID:      token.TeamID,
		CreatedAt:   sql.Timestamptz(token.CreatedAt),
		Expiry:      sql.TimestamptzPtr(token.Expiry),
	})
	return err
}

func (db *pgdb) getTeamTokenByTeamID(ctx context.Context, teamID resource.TfeID) (*Token, error) {
	// query only returns 0 or 1 tokens
	result, err := q.FindTeamTokensByID(ctx, db.Conn(ctx), teamID)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	ot := &Token{
		TfeID:     result[0].TeamTokenID,
		CreatedAt: result[0].CreatedAt.Time.UTC(),
		TeamID:    result[0].TeamID,
	}
	if result[0].Expiry.Valid {
		ot.Expiry = internal.Time(result[0].Expiry.Time.UTC())
	}
	return ot, nil
}

func (db *pgdb) deleteTeamToken(ctx context.Context, teamID resource.TfeID) error {
	_, err := q.DeleteTeamTokenByID(ctx, db.Conn(ctx), teamID)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
