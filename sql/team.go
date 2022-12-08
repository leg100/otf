package sql

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateTeam persists a team to the DB.
func (db *DB) CreateTeam(ctx context.Context, team *otf.Team) error {
	_, err := db.InsertTeam(ctx, pggen.InsertTeamParams{
		ID:             String(team.ID()),
		Name:           String(team.Name()),
		CreatedAt:      Timestamptz(team.CreatedAt()),
		OrganizationID: String(team.Organization().ID()),
	})
	return err
}

func (db *DB) UpdateTeam(ctx context.Context, name, organization string, fn func(*otf.Team) error) (*otf.Team, error) {
	var team *otf.Team
	err := db.tx(ctx, func(tx *DB) error {
		var err error

		// retrieve team
		result, err := tx.FindTeamByNameForUpdate(ctx, String(name), String(organization))
		if err != nil {
			return err
		}
		team = otf.UnmarshalTeamResult(otf.TeamResult(result))

		// update team
		if err := fn(team); err != nil {
			return err
		}
		// persist update
		_, err = tx.UpdateTeamByName(ctx, pggen.UpdateTeamByNameParams{
			PermissionManageWorkspaces: team.OrganizationAccess().ManageWorkspaces,
			PermissionManageVCS:        team.OrganizationAccess().ManageVCS,
			OrganizationName:           String(organization),
			Name:                       String(name),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return team, err
}

// GetTeam retrieves a team from the DB
func (db *DB) GetTeam(ctx context.Context, name, organization string) (*otf.Team, error) {
	result, err := db.FindTeamByName(ctx, String(name), String(organization))
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalTeamResult(otf.TeamResult(result)), nil
}

func (db *DB) ListTeams(ctx context.Context, organization string) ([]*otf.Team, error) {
	result, err := db.FindTeamsByOrg(ctx, String(organization))
	if err != nil {
		return nil, err
	}

	var items []*otf.Team
	for _, r := range result {
		items = append(items, otf.UnmarshalTeamResult(otf.TeamResult(r)))
	}
	return items, nil
}

func (db *DB) AddTeamMembership(ctx context.Context, userID, teamID string) error {
	_, err := db.InsertTeamMembership(ctx, String(userID), String(teamID))
	if err != nil {
		return databaseError(err)
	}
	return nil
}

func (db *DB) RemoveTeamMembership(ctx context.Context, userID, teamID string) error {
	_, err := db.DeleteTeamMembership(ctx, String(userID), String(teamID))
	if err != nil {
		return databaseError(err)
	}
	return nil
}

// DeleteTeam deletes a team from the DB.
func (db *DB) DeleteTeam(ctx context.Context, name, organization string) error {
	_, err := db.DeleteTeamByName(ctx, String(name), String(organization))
	if err != nil {
		return databaseError(err)
	}
	return nil
}
