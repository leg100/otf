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

// EnsureCreatedTeam retrieves the team or creates the team if it doesn't exist.
func (a *Application) EnsureCreatedTeam(ctx context.Context, name, organizationName string) (*otf.Team, error) {
	team, err := a.db.GetTeam(ctx, otf.TeamSpec{
		Name:             otf.String(name),
		OrganizationName: otf.String(organizationName),
	})
	if err == nil {
		return team, nil
	}
	if err != otf.ErrResourceNotFound {
		a.Error(err, "retrieving team", "name", name, "organization", organizationName)
		return nil, err
	}

	return a.CreateTeam(ctx, name, organizationName)
}

func (a *Application) GetTeam(ctx context.Context, spec otf.TeamSpec) (*otf.Team, error) {
	team, err := a.db.GetTeam(ctx, spec)
	if err != nil {
		a.V(2).Info("retrieving team", spec.KeyValue()...)
		return nil, err
	}
	a.V(2).Info("retrieved team", "name", team.Name(), "organization", team.OrganizationName())

	return team, nil
}

func (a *Application) ListTeams(ctx context.Context, organizationName string) ([]*otf.Team, error) {
	teams, err := a.db.ListTeams(ctx, organizationName)
	if err != nil {
		a.V(2).Info("listing teams", "organization", organizationName)
		return nil, err
	}
	a.V(2).Info("listed teams", "organization", organizationName, "total", len(teams))

	return teams, nil
}
