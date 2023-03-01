package auth

import "context"

type fakeTeamApp struct {
	team    *Team
	members []*User

	teamService
}

func (f *fakeTeamApp) getTeamByID(ctx context.Context, teamID string) (*Team, error) {
	return f.team, nil
}

func (f *fakeTeamApp) listTeams(ctx context.Context, organization string) ([]*Team, error) {
	return []*Team{f.team}, nil
}

func (f *fakeTeamApp) updateTeam(ctx context.Context, teamID string, opts UpdateTeamOptions) (*Team, error) {
	f.team.Update(opts)
	return f.team, nil
}

func (f *fakeTeamApp) listTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
	return f.members, nil
}
