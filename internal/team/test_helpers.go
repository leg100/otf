package team

import (
	"context"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

type fakeService struct {
	team *Team
}

func (f *fakeService) Create(context.Context, organization.Name, CreateTeamOptions) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) Update(context.Context, resource.TfeID, UpdateTeamOptions) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) Get(context.Context, organization.Name, string) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) GetByID(context.Context, resource.TfeID) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) List(context.Context, organization.Name) ([]*Team, error) {
	return []*Team{f.team}, nil
}

func (f *fakeService) Delete(context.Context, resource.TfeID) error {
	return nil
}
