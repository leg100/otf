package team

import (
	"context"

	"github.com/leg100/otf/internal/resource"
)

type fakeService struct {
	team *Team
}

func (f *fakeService) Create(context.Context, resource.OrganizationName, CreateTeamOptions) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) Update(context.Context, resource.ID, UpdateTeamOptions) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) Get(context.Context, resource.OrganizationName, string) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) GetByID(context.Context, resource.ID) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) List(context.Context, resource.OrganizationName) ([]*Team, error) {
	return []*Team{f.team}, nil
}

func (f *fakeService) Delete(context.Context, resource.ID) error {
	return nil
}
