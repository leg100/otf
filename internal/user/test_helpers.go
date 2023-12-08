package user

import (
	"context"

	"github.com/leg100/otf/internal/team"
)

type fakeService struct {
	user  *User
	token []byte
	ut    *UserToken

	*Service
}

func (f *fakeService) Create(context.Context, string, ...NewUserOption) (*User, error) {
	return f.user, nil
}

func (f *fakeService) List(ctx context.Context) ([]*User, error) {
	return []*User{f.user}, nil
}

func (f *fakeService) ListTeamUsers(ctx context.Context, teamID string) ([]*User, error) {
	return []*User{f.user}, nil
}

func (f *fakeService) Delete(context.Context, string) error {
	return nil
}

func (f *fakeService) AddTeamMembership(context.Context, string, []string) error {
	return nil
}

func (f *fakeService) RemoveTeamMembership(context.Context, string, []string) error {
	return nil
}

func (f *fakeService) CreateToken(context.Context, CreateUserTokenOptions) (*UserToken, []byte, error) {
	return nil, f.token, nil
}

func (f *fakeService) ListTokens(context.Context) ([]*UserToken, error) {
	return []*UserToken{f.ut}, nil
}

func (f *fakeService) DeleteToken(context.Context, string) error {
	return nil
}

type fakeTeamService struct {
	team *team.Team
}

func (f *fakeTeamService) Get(context.Context, string, string) (*team.Team, error) {
	return f.team, nil
}

func (f *fakeTeamService) GetByID(context.Context, string) (*team.Team, error) {
	return f.team, nil
}
