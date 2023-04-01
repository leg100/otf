package authenticator

import (
	"context"
	"net/http"

	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"golang.org/x/oauth2"
)

type fakeAuthenticatorService struct {
	auth.AuthService
}

func (f *fakeAuthenticatorService) sync(context.Context, cloud.User) (*auth.User, error) {
	return auth.NewUser("fake-user"), nil
}

func (f *fakeAuthenticatorService) CreateSession(context.Context, auth.CreateSessionOptions) (*auth.Session, error) {
	return &auth.Session{}, nil
}

type fakeOAuthClient struct {
	user *cloud.User
	oauthClient
}

func (f *fakeOAuthClient) CallbackHandler(*http.Request) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func (f *fakeOAuthClient) NewClient(context.Context, *oauth2.Token) (cloud.Client, error) {
	return &fakeCloudClient{user: f.user}, nil
}

type fakeCloudClient struct {
	user *cloud.User
	cloud.Client
}

func (f *fakeCloudClient) GetUser(context.Context) (*cloud.User, error) {
	return f.user, nil
}

type fakeAuthService struct {
	addedTeams, removedTeams []string

	auth.AuthService
}

func (f *fakeAuthService) AddTeamMembership(ctx context.Context, _, team string) error {
	f.addedTeams = append(f.addedTeams, team)
	return nil
}

func (f *fakeAuthService) RemoveTeamMembership(ctx context.Context, _, team string) error {
	f.removedTeams = append(f.removedTeams, team)
	return nil
}

type fakeUserSynchroniser struct{}

func (f *fakeUserSynchroniser) sync(ctx context.Context, from cloud.User) (*auth.User, error) {
	return &auth.User{}, nil
}
