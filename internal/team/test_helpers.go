package team

import (
	"context"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

type fakeService struct {
	team *Team
}

func (f *fakeService) CreateTeam(context.Context, organization.Name, CreateTeamOptions) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) UpdateTeam(context.Context, resource.ID, UpdateTeamOptions) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) GetTeam(context.Context, organization.Name, string) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) GetTeamByID(context.Context, resource.ID) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) ListTeams(context.Context, organization.Name) ([]*Team, error) {
	return []*Team{f.team}, nil
}

func (f *fakeService) DeleteTeam(context.Context, resource.ID) error {
	return nil
}
