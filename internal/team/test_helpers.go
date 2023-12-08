package team

import "context"

type fakeService struct {
	team *Team
}

func (f *fakeService) Create(context.Context, string, CreateTeamOptions) (*Team, error) {
	return f.team, nil

}

func (f *fakeService) Update(context.Context, string, UpdateTeamOptions) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) Get(context.Context, string, string) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) GetByID(context.Context, string) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) List(context.Context, string) ([]*Team, error) {
	return []*Team{f.team}, nil
}

func (f *fakeService) Delete(context.Context, string) error {
	return nil
}
