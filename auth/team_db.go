package auth

import (
	"context"

	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

func (db *pgdb) createTeam(ctx context.Context, team *Team) error {
	_, err := db.InsertTeam(ctx, pggen.InsertTeamParams{
		ID:               sql.String(team.ID),
		Name:             sql.String(team.Name),
		CreatedAt:        sql.Timestamptz(team.CreatedAt),
		OrganizationName: sql.String(team.Organization),
	})
	return sql.Error(err)
}

func (db *pgdb) UpdateTeam(ctx context.Context, teamID string, fn func(*Team) error) (*Team, error) {
	var team *Team
	err := db.tx(ctx, func(tx *pgdb) error {
		var err error

		// retrieve team
		result, err := tx.FindTeamByIDForUpdate(ctx, sql.String(teamID))
		if err != nil {
			return err
		}
		team = teamRow(result).toTeam()

		// update team
		if err := fn(team); err != nil {
			return err
		}
		// persist update
		_, err = tx.UpdateTeamByID(ctx, pggen.UpdateTeamByIDParams{
			PermissionManageWorkspaces: team.OrganizationAccess().ManageWorkspaces,
			PermissionManageVCS:        team.OrganizationAccess().ManageVCS,
			PermissionManageRegistry:   team.OrganizationAccess().ManageRegistry,
			TeamID:                     sql.String(teamID),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return team, err
}

func (db *pgdb) getTeam(ctx context.Context, name, organization string) (*Team, error) {
	result, err := db.FindTeamByName(ctx, sql.String(name), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return teamRow(result).toTeam(), nil
}

func (db *pgdb) getTeamByID(ctx context.Context, id string) (*Team, error) {
	result, err := db.FindTeamByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return teamRow(result).toTeam(), nil
}

func (db *pgdb) listTeams(ctx context.Context, organization string) ([]*Team, error) {
	result, err := db.FindTeamsByOrg(ctx, sql.String(organization))
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
	_, err := db.DeleteTeamByID(ctx, sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
