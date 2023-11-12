package user

import (
	"context"

	"github.com/leg100/otf/internal/team"
)

type fakeService struct {
	user  *User
	token []byte
	ut    *UserToken

	UserService
}

func (f *fakeService) CreateUser(context.Context, string, ...NewUserOption) (*User, error) {
	return f.user, nil
}

func (f *fakeService) ListUsers(ctx context.Context) ([]*User, error) {
	return []*User{f.user}, nil
}

func (f *fakeService) ListTeamUsers(ctx context.Context, teamID string) ([]*User, error) {
	return []*User{f.user}, nil
}

func (f *fakeService) DeleteUser(context.Context, string) error {
	return nil
}

func (f *fakeService) AddTeamMembership(context.Context, string, []string) error {
	return nil
}

func (f *fakeService) RemoveTeamMembership(context.Context, string, []string) error {
	return nil
}

func (f *fakeService) CreateUserToken(context.Context, CreateUserTokenOptions) (*UserToken, []byte, error) {
	return nil, f.token, nil
}

func (f *fakeService) ListUserTokens(context.Context) ([]*UserToken, error) {
	return []*UserToken{f.ut}, nil
}

func (f *fakeService) DeleteUserToken(context.Context, string) error {
	return nil
}

type fakeTeamService struct {
	team *team.Team

	team.TeamService
}

func (f *fakeTeamService) GetTeam(context.Context, string, string) (*team.Team, error) {
	return f.team, nil
}

func (f *fakeTeamService) GetTeamByID(context.Context, string) (*team.Team, error) {
	return f.team, nil
}
