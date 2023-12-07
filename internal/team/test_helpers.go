package team

import "context"

type fakeService struct {
	team *Team
}

func (f *fakeService) CreateTeam(context.Context, string, CreateTeamOptions) (*Team, error) {
	return f.team, nil

}

func (f *fakeService) UpdateTeam(context.Context, string, UpdateTeamOptions) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) GetTeam(context.Context, string, string) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) GetTeamByID(context.Context, string) (*Team, error) {
	return f.team, nil
}

func (f *fakeService) ListTeams(context.Context, string) ([]*Team, error) {
	return []*Team{f.team}, nil
}

func (f *fakeService) DeleteTeam(context.Context, string) error {
	return nil
}
