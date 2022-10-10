package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateTeam persists a team to the DB.
func (db *DB) CreateTeam(ctx context.Context, team *otf.Team) error {
	_, err := db.InsertTeam(ctx, pggen.InsertTeamParams{
		ID:             String(team.ID()),
		Name:           String(team.Name()),
		CreatedAt:      Timestamptz(team.CreatedAt()),
		OrganizationID: String(team.OrganizationID()),
	})
	return err
}

func (db *DB) UpdateTeam(ctx context.Context, spec otf.TeamSpec, fn func(*otf.Team) error) (*otf.Team, error) {
	var team *otf.Team
	err := db.tx(ctx, func(tx *DB) error {
		var err error
		// retrieve team
		team, err = tx.getTeamForUpdate(ctx, spec)
		if err != nil {
			return databaseError(err)
		}
		// update team
		if err := fn(team); err != nil {
			return err
		}
		// persist update
		_, err = tx.UpdateTeamByID(ctx, team.OrganizationAccess().ManageWorkspaces, String(team.ID()))
		if err != nil {
			return err
		}
		return nil
	})
	return team, err
}

// GetTeam retrieves a team from the DB
func (db *DB) GetTeam(ctx context.Context, spec otf.TeamSpec) (*otf.Team, error) {
	if spec.ID != nil {
		result, err := db.FindTeamByID(ctx, String(*spec.ID))
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalTeamDBResult(otf.TeamDBResult(result)), nil
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := db.FindTeamByName(ctx, String(*spec.Name), String(*spec.OrganizationName))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalTeamDBResult(otf.TeamDBResult(result)), nil
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func (db *DB) ListTeams(ctx context.Context, organizationName string) ([]*otf.Team, error) {
	result, err := db.FindTeamsByOrg(ctx, String(organizationName))
	if err != nil {
		return nil, err
	}

	var items []*otf.Team
	for _, r := range result {
		items = append(items, otf.UnmarshalTeamDBResult(otf.TeamDBResult(r)))
	}
	return items, nil
}

func (db *DB) AddTeamMembership(ctx context.Context, userID, teamID string) error {
	_, err := db.InsertTeamMembership(ctx, String(userID), String(teamID))
	return err
}

func (db *DB) RemoveTeamMembership(ctx context.Context, userID, teamID string) error {
	_, err := db.DeleteTeamMembership(ctx, String(userID), String(teamID))
	if err != nil {
		return databaseError(err)
	}
	return nil
}

// DeleteTeam deletes a team from the DB.
func (db *DB) DeleteTeam(ctx context.Context, spec otf.TeamSpec) error {
	if spec.ID != nil {
		_, err := db.DeleteTeamByID(ctx, String(*spec.ID))
		if err != nil {
			return databaseError(err)
		}
	} else if spec.Name != nil && spec.OrganizationName != nil {
		_, err := db.DeleteTeamByName(ctx, String(*spec.Name), String(*spec.OrganizationName))
		if err != nil {
			return databaseError(err)
		}
	} else {
		return fmt.Errorf("unsupported team spec for deletion")
	}
	return nil
}

func (db *DB) getTeamForUpdate(ctx context.Context, spec otf.TeamSpec) (*otf.Team, error) {
	id, err := db.getTeamID(ctx, spec)
	if err != nil {
		return nil, err
	}
	result, err := db.FindTeamByIDForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	return otf.UnmarshalTeamDBResult(otf.TeamDBResult(result)), nil
}

func (db *DB) getTeamID(ctx context.Context, spec otf.TeamSpec) (pgtype.Text, error) {
	if spec.ID != nil {
		return String(*spec.ID), nil
	}
	if spec.Name != nil && spec.OrganizationName != nil {
		return db.FindTeamIDByName(ctx, String(*spec.Name), String(*spec.OrganizationName))
	}
	return pgtype.Text{}, otf.ErrInvalidTeamSpec
}
