package auth

import (
	"context"

	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

func (db *pgdb) createTeam(ctx context.Context, team *Team) error {
	_, err := db.Conn(ctx).InsertTeam(ctx, pggen.InsertTeamParams{
		ID:                              sql.String(team.ID),
		Name:                            sql.String(team.Name),
		CreatedAt:                       sql.Timestamptz(team.CreatedAt),
		OrganizationName:                sql.String(team.Organization),
		Visibility:                      sql.String(team.Visibility),
		SSOTeamID:                       sql.StringPtr(team.SSOTeamID),
		PermissionManageWorkspaces:      team.Access.ManageWorkspaces,
		PermissionManageVCS:             team.Access.ManageVCS,
		PermissionManageModules:         team.Access.ManageModules,
		PermissionManageProviders:       team.Access.ManageProviders,
		PermissionManagePolicies:        team.Access.ManagePolicies,
		PermissionManagePolicyOverrides: team.Access.ManagePolicyOverrides,
	})
	return sql.Error(err)
}

func (db *pgdb) UpdateTeam(ctx context.Context, teamID string, fn func(*Team) error) (*Team, error) {
	var team *Team
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		var err error

		// retrieve team
		result, err := q.FindTeamByIDForUpdate(ctx, sql.String(teamID))
		if err != nil {
			return err
		}
		team = teamRow(result).toTeam()

		// update team
		if err := fn(team); err != nil {
			return err
		}
		// persist update
		_, err = q.UpdateTeamByID(ctx, pggen.UpdateTeamByIDParams{
			TeamID:                          sql.String(teamID),
			Name:                            sql.String(team.Name),
			Visibility:                      sql.String(team.Visibility),
			SSOTeamID:                       sql.StringPtr(team.SSOTeamID),
			PermissionManageWorkspaces:      team.Access.ManageWorkspaces,
			PermissionManageVCS:             team.Access.ManageVCS,
			PermissionManageModules:         team.Access.ManageModules,
			PermissionManageProviders:       team.Access.ManageProviders,
			PermissionManagePolicies:        team.Access.ManagePolicies,
			PermissionManagePolicyOverrides: team.Access.ManagePolicyOverrides,
		})
		if err != nil {
			return err
		}
		return nil
	})
	return team, err
}

func (db *pgdb) getTeam(ctx context.Context, name, organization string) (*Team, error) {
	result, err := db.Conn(ctx).FindTeamByName(ctx, sql.String(name), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return teamRow(result).toTeam(), nil
}

func (db *pgdb) getTeamByID(ctx context.Context, id string) (*Team, error) {
	result, err := db.Conn(ctx).FindTeamByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return teamRow(result).toTeam(), nil
}

func (db *pgdb) getTeamForUpdate(ctx context.Context, id string) (*Team, error) {
	result, err := db.Conn(ctx).FindTeamByIDForUpdate(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return teamRow(result).toTeam(), nil
}

func (db *pgdb) listTeams(ctx context.Context, organization string) ([]*Team, error) {
	result, err := db.Conn(ctx).FindTeamsByOrg(ctx, sql.String(organization))
	if err != nil {
		return nil, err
	}

	var items []*Team
	for _, r := range result {
		items = append(items, teamRow(r).toTeam())
	}
	return items, nil
}

func (db *pgdb) deleteTeam(ctx context.Context, teamID string) error {
	_, err := db.Conn(ctx).DeleteTeamByID(ctx, sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
