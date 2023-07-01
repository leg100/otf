package auth

import (
	"context"
	"errors"

	"github.com/leg100/otf/internal/rbac"
)

var ErrRemovingOwnersTeamNotPermitted = errors.New("the owners team cannot be deleted")

type TeamService interface {
	CreateTeam(ctx context.Context, organization string, opts CreateTeamOptions) (*Team, error)
	GetTeam(ctx context.Context, organization, team string) (*Team, error)
	GetTeamByID(ctx context.Context, teamID string) (*Team, error)
	ListTeams(ctx context.Context, organization string) ([]*Team, error)
	ListTeamMembers(ctx context.Context, teamID string) ([]*User, error)
	UpdateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error)
	DeleteTeam(ctx context.Context, teamID string) error
}

// CreateTeam creates a team. If Tx in opts is non-nil then the team is created
// within that database transaction.
func (a *service) CreateTeam(ctx context.Context, organization string, opts CreateTeamOptions) (*Team, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateTeamAction, organization)
	if err != nil {
		return nil, err
	}

	team, err := newTeam(organization, opts)
	if err != nil {
		return nil, err
	}

	db := a.db
	if opts.Tx != nil {
		db = newDB(opts.Tx, a.db.Logger)
	}
	if err := db.createTeam(ctx, team); err != nil {
		a.Error(err, "creating team", "name", opts.Name, "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created team", "name", opts.Name, "organization", organization, "subject", subject)

	return team, nil
}

func (a *service) UpdateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error) {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}
	subject, err := a.organization.CanAccess(ctx, rbac.UpdateTeamAction, team.Organization)
	if err != nil {
		return nil, err
	}

	team, err = a.db.UpdateTeam(ctx, teamID, func(team *Team) error {
		return team.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating team", "name", team.Name, "organization", team.Organization, "subject", subject)
		return nil, err
	}

	a.V(2).Info("updated team", "name", team.Name, "organization", team.Organization, "subject", subject)

	return team, nil
}

// ListTeams lists teams in the organization.
func (a *service) ListTeams(ctx context.Context, organization string) ([]*Team, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListTeamsAction, organization)
	if err != nil {
		return nil, err
	}

	teams, err := a.db.listTeams(ctx, organization)
	if err != nil {
		a.Error(err, "listing teams", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(9).Info("listed teams", "organization", organization, "subject", subject)

	return teams, nil
}

// ListTeamMembers lists users that are members of the given team. The caller
// needs either organization-wide authority to call this endpoint, or they need
// to be a member of the team.
func (a *service) ListTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.ListUsersAction, team.Organization)
	if err != nil {
		return nil, err
	}

	members, err := a.db.listTeamMembers(ctx, teamID)
	if err != nil {
		a.Error(err, "listing team members", "team_id", teamID, "subject", subject)
		return nil, err
	}

	a.V(9).Info("listed team members", "team_id", teamID, "subject", subject)

	return members, nil
}

func (a *service) GetTeam(ctx context.Context, organization, name string) (*Team, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.GetTeamAction, organization)
	if err != nil {
		return nil, err
	}

	team, err := a.db.getTeam(ctx, name, organization)
	if err != nil {
		a.Error(err, "retrieving team", "team", name, "organization", organization, "subject", subject)
		return nil, err
	}

	a.V(9).Info("retrieved team", "team", name, "organization", organization, "subject", subject)

	return team, nil
}

func (a *service) GetTeamByID(ctx context.Context, teamID string) (*Team, error) {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.GetTeamAction, team.Organization)
	if err != nil {
		return nil, err
	}

	a.V(9).Info("retrieved team", "team", team.Name, "organization", team.Organization, "subject", subject)

	return team, nil
}

func (a *service) DeleteTeam(ctx context.Context, teamID string) error {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteTeamAction, team.Organization)
	if err != nil {
		return err
	}

	if team.Name == "owners" {
		return ErrRemovingOwnersTeamNotPermitted
	}

	err = a.db.deleteTeam(ctx, teamID)
	if err != nil {
		a.Error(err, "deleting team", "team", team.Name, "organization", team.Organization, "subject", subject)
		return err
	}

	a.V(2).Info("deleted team", "team", team.Name, "organization", team.Organization, "subject", subject)

	return nil
}
