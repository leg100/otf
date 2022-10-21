package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CreateTeam(ctx context.Context, name, organization string) (*otf.Team, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.CreateTeamAction, organization)
	if err != nil {
		return nil, err
	}

	org, err := a.db.GetOrganization(ctx, organization)
	if err != nil {
		return nil, err
	}

	team := otf.NewTeam(name, org)

	if err := a.db.CreateTeam(ctx, team); err != nil {
		a.Error(err, "creating team", "name", name, "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created team", "name", name, "organization", organization, "subject", subject)

	return team, nil
}

func (a *Application) UpdateTeam(ctx context.Context, name, organization string, opts otf.TeamUpdateOptions) (*otf.Team, error) {
	subject, err := a.CanAccessOrganization(ctx, otf.UpdateTeamAction, organization)
	if err != nil {
		return nil, err
	}

	team, err := a.db.UpdateTeam(ctx, name, organization, func(team *otf.Team) error {
		return team.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating team", "name", name, "organization", organization, "subject", subject)
		return nil, err
	}

	a.V(2).Info("updated team", "name", name, "organization", organization, "subject", subject)

	return team, nil
}

// EnsureCreatedTeam retrieves the team or creates the team if it doesn't exist.
func (a *Application) EnsureCreatedTeam(ctx context.Context, name, organization string) (*otf.Team, error) {
	team, err := a.GetTeam(ctx, name, organization)
	if err == otf.ErrResourceNotFound {
		return a.CreateTeam(ctx, name, organization)
	} else if err != nil {
		a.Error(err, "retrieving team", "name", name, "organization", organization)
		return nil, err
	} else {
		return team, nil
	}
}

// GetTeam retrieves a team in an organization. If the caller is an unprivileged
// user i.e. not an owner nor a site admin then they are only permitted to
// retrieve a team they are a member of.
func (a *Application) GetTeam(ctx context.Context, name, organization string) (*otf.Team, error) {
	if user, err := otf.UserFromContext(ctx); err == nil && user.IsUnprivilegedUser(organization) {
		team, err := user.Team(name, organization)
		if err != nil {
			return nil, otf.ErrAccessNotPermitted
		}
		a.V(2).Info("retrieved team", "organization", organization, "team", name, "subject", user)
		return team, nil
	}

	subject, err := a.CanAccessOrganization(ctx, otf.GetTeamAction, organization)
	if err != nil {
		return nil, err
	}

	team, err := a.db.GetTeam(ctx, name, organization)
	if err != nil {
		a.V(2).Info("retrieving team", "team", name, "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved team", "team", name, "organization", organization, "subject", subject)

	return team, nil
}

// ListTeams lists teams in the organization. If the caller is an unprivileged
// user i.e. not an owner nor a site admin then only their teams are listed.
func (a *Application) ListTeams(ctx context.Context, organization string) ([]*otf.Team, error) {
	if user, err := otf.UserFromContext(ctx); err == nil && user.IsUnprivilegedUser(organization) {
		a.V(2).Info("listed teams", "organization", organization, "subject", user)
		return user.TeamsByOrganization(organization), nil
	}

	subject, err := a.CanAccessOrganization(ctx, otf.ListTeamsAction, organization)
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
