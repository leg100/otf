package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CreateTeam(ctx context.Context, name, organizationName string) (*otf.Team, error) {
	org, err := a.db.GetOrganization(ctx, organizationName)
	if err != nil {
		return nil, err
	}

	team := otf.NewTeam(name, org)

	if err := a.db.CreateTeam(ctx, team); err != nil {
		a.Error(err, "creating team", "name", name, "organization", organizationName)
		return nil, err
	}
	a.V(0).Info("created team", "name", name, "organization", organizationName)

	return team, nil
}

func (a *Application) UpdateTeam(ctx context.Context, name, organization string, opts otf.TeamUpdateOptions) (*otf.Team, error) {
	team, err := a.db.UpdateTeam(ctx, name, organization, func(team *otf.Team) error {
		return team.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating team", "name", name, "organization", organization)
		return nil, err
	}

	a.V(2).Info("updated team", "name", name, "organization", organization)

	return team, nil
}

// EnsureCreatedTeam retrieves the team or creates the team if it doesn't exist.
func (a *Application) EnsureCreatedTeam(ctx context.Context, name, organization string) (*otf.Team, error) {
	team, err := a.db.GetTeam(ctx, name, organization)
	if err == nil {
		return team, nil
	}
	if err != otf.ErrResourceNotFound {
		a.Error(err, "retrieving team", "name", name, "organization", organization)
		return nil, err
	}

	return a.CreateTeam(ctx, name, organization)
}

func (a *Application) GetTeam(ctx context.Context, name, organization string) (*otf.Team, error) {
	team, err := a.db.GetTeam(ctx, name, organization)
	if err != nil {
		a.V(2).Info("retrieving team", "name", name, "organization", organization)
		return nil, err
	}
	a.V(2).Info("retrieved team", "name", name, "organization", organization)

	return team, nil
}

// ListTeams lists teams in the organization. If the caller is a normal user then only their teams
// are listed.
func (a *Application) ListTeams(ctx context.Context, organization string) ([]*otf.Team, error) {
	subject, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if user, ok := subject.(*otf.User); ok && !user.IsSiteAdmin() {
		a.V(2).Info("listed teams", "organization", organization, "subject", subject)
		return user.TeamsByOrganization(organization), nil
	}

	teams, err := a.db.ListTeams(ctx, organization)
	if err != nil {
		a.V(2).Info("listing teams", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed teams", "organization", organization, "subject", subject)

	return teams, nil
}
