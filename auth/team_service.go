package auth

import (
	"context"

	"github.com/leg100/otf/rbac"
)

type teamService interface {
	createTeam(ctx context.Context, opts NewTeamOptions) (*Team, error)
	getTeam(ctx context.Context, organization, team string) (*Team, error)
	getTeamByID(ctx context.Context, teamID string) (*Team, error)
	listTeams(ctx context.Context, organization string) ([]*Team, error)
	listTeamMembers(ctx context.Context, teamID string) ([]*User, error)
	updateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error)
	deleteTeam(ctx context.Context, teamID string) error
}

func (a *service2) CreateTeam(ctx context.Context, opts NewTeamOptions) (*Team, error) {
	return a.createTeam(ctx, opts)
}

func (a *service2) UpdateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error) {
	return a.updateTeam(ctx, teamID, opts)
}

func (a *service2) ListTeams(ctx context.Context, organization string) ([]*Team, error) {
	return a.listTeams(ctx, organization)
}

func (a *service2) ListTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
	return a.listTeamMembers(ctx, teamID)
}

func (a *service2) GetTeam(ctx context.Context, organization, name string) (*Team, error) {
	return a.getTeam(ctx, organization, name)
}

func (a *service2) GetTeamByID(ctx context.Context, teamID string) (*Team, error) {
	return a.getTeamByID(ctx, teamID)
}

func (a *service2) DeleteTeam(ctx context.Context, teamID string) error {
	return a.deleteTeam(ctx, teamID)
}

func (a *service2) createTeam(ctx context.Context, opts NewTeamOptions) (*Team, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateTeamAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	team := NewTeam(opts)

	if err := a.db.createTeam(ctx, team); err != nil {
		a.Error(err, "creating team", "name", opts.Name, "organization", opts.Organization, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created team", "name", opts.Name, "organization", opts.Organization, "subject", subject)

	return team, nil
}

func (a *service2) updateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error) {
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

func (a *service2) getTeam(ctx context.Context, organization, name string) (*Team, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.GetTeamAction, organization)
	if err != nil {
		return nil, err
	}

	team, err := a.db.getTeam(ctx, name, organization)
	if err != nil {
		a.Error(err, "retrieving team", "team", name, "organization", organization, "subject", subject)
		return nil, err
	}

	a.V(2).Info("retrieved team", "team", name, "organization", organization, "subject", subject)

	return team, nil
}

func (a *service2) getTeamByID(ctx context.Context, teamID string) (*Team, error) {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.GetTeamAction, team.Organization)
	if err != nil {
		return nil, err
	}

	a.V(2).Info("retrieved team", "team", team.Name, "organization", team.Organization, "subject", subject)

	return team, nil
}

// listTeams lists teams in the organization.
func (a *service2) listTeams(ctx context.Context, organization string) ([]*Team, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListTeamsAction, organization)
	if err != nil {
		return nil, err
	}

	teams, err := a.db.listTeams(ctx, organization)
	if err != nil {
		a.V(2).Info("listing teams", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed teams", "organization", organization, "subject", subject)

	return teams, nil
}

// listTeamMembers lists users that are members of the given team. The caller
// needs either organization-wide authority to call this endpoint, or they need
// to be a member of the team.
func (a *service2) listTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
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

	a.V(2).Info("listed team members", "team_id", teamID, "subject", subject)

	return members, nil
}

func (a *service2) deleteTeam(ctx context.Context, teamID string) error {
	team, err := a.db.getTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.GetTeamAction, team.Organization)
	if err != nil {
		return err
	}

	err = a.db.deleteTeam(ctx, teamID)
	if err != nil {
		a.Error(err, "deleting team", "team", team.Name, "organization", team.Organization, "subject", subject)
		return err
	}

	a.V(2).Info("deleted team", "team", team.Name, "organization", team.Organization, "subject", subject)

	return nil
}
