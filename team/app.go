package team

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

func (a *Application) CreateTeam(ctx context.Context, opts otf.CreateTeamOptions) (*otf.Team, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.CreateTeamAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	org, err := a.db.GetOrganization(ctx, opts.Organization)
	if err != nil {
		return nil, err
	}

	team := otf.NewTeam(opts.Name, org)

	if err := a.db.CreateTeam(ctx, team); err != nil {
		a.Error(err, "creating team", "name", opts.Name, "organization", opts.Organization, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created team", "name", opts.Name, "organization", opts.Organization, "subject", subject)

	return team, nil
}

// EnsureCreatedTeam retrieves the team or creates the team if it doesn't exist.
func (a *Application) EnsureCreatedTeam(ctx context.Context, opts otf.CreateTeamOptions) (*otf.Team, error) {
	team, err := a.db.GetTeam(ctx, opts.Name, opts.Organization)
	if err == otf.ErrResourceNotFound {
		return a.CreateTeam(ctx, opts)
	} else if err != nil {
		a.Error(err, "retrieving team", "name", opts.Name, "organization", opts.Organization)
		return nil, err
	} else {
		return team, nil
	}
}

func (a *Application) UpdateTeam(ctx context.Context, teamID string, opts otf.UpdateTeamOptions) (*otf.Team, error) {
	team, err := a.db.GetTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}
	subject, err := a.CanAccessOrganization(ctx, rbac.UpdateTeamAction, team.Organization())
	if err != nil {
		return nil, err
	}

	team, err = a.db.UpdateTeam(ctx, teamID, func(team *otf.Team) error {
		return team.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating team", "name", team.Name(), "organization", team.Organization(), "subject", subject)
		return nil, err
	}

	a.V(2).Info("updated team", "name", team.Name(), "organization", team.Organization(), "subject", subject)

	return team, nil
}

// GetTeam retrieves a team in an organization. If the caller is an unprivileged
// user i.e. not an owner nor a site admin then they are only permitted to
// retrieve a team they are a member of.
func (a *Application) GetTeam(ctx context.Context, teamID string) (*otf.Team, error) {
	team, err := a.db.GetTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	// Check organization-wide authority
	subject, err := a.CanAccessOrganization(ctx, rbac.GetTeamAction, team.Organization())
	if err != nil {
		// Fallback to checking if they are member of the team
		if user, ok := subject.(*otf.User); ok {
			if !user.IsTeamMember(teamID) {
				// User is not a member; refuse access
				return nil, err
			}
		} else {
			// non-user without organization-wide authority; refuse access
			return nil, err
		}
	}

	a.V(2).Info("retrieved team", "team", team.Name(), "organization", team.Organization(), "subject", subject)

	return team, nil
}

// ListTeams lists teams in the organization. If the caller is an unprivileged
// user i.e. not an owner nor a site admin then only their teams are listed.
func (a *Application) ListTeams(ctx context.Context, organization string) ([]*otf.Team, error) {
	if user, err := otf.UserFromContext(ctx); err == nil && user.IsUnprivilegedUser(organization) {
		a.V(2).Info("listed teams", "organization", organization, "subject", user)
		return user.TeamsByOrganization(organization), nil
	}

	subject, err := a.CanAccessOrganization(ctx, rbac.ListTeamsAction, organization)
	if err != nil {
		return nil, err
	}

	teams, err := a.db.ListTeams(ctx, organization)
	if err != nil {
		a.V(2).Info("listing teams", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed teams", "organization", organization, "subject", subject)

	return teams, nil
}

// ListTeamMembers lists users that are members of the given team. The caller
// needs either organization-wide authority to call this endpoint, or they need
// to be a member of the team.
func (a *Application) ListTeamMembers(ctx context.Context, teamID string) ([]*otf.User, error) {
	team, err := a.db.GetTeamByID(ctx, teamID)
	if err != nil {
		a.Error(err, "retrieving team", "team_id", teamID)
		return nil, err
	}

	// Check organization-wide authority
	subject, err := a.CanAccessOrganization(ctx, rbac.ListTeamsAction, team.Organization())
	if err != nil {
		// Fallback to checking if they are member of the team
		if user, ok := subject.(*otf.User); ok {
			if !user.IsTeamMember(teamID) {
				// User is not a member; refuse access
				return nil, err
			}
		} else {
			// non-user without organization-wide authority; refuse access
			return nil, err
		}
	}

	members, err := a.db.ListTeamMembers(ctx, teamID)
	if err != nil {
		a.Error(err, "listing team members", "team_id", teamID, "subject", subject)
		return nil, err
	}

	a.V(2).Info("listed team members", "team_id", teamID, "subject", subject)

	return members, nil
}
