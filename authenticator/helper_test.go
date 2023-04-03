package authenticator

import (
	"context"
	"net/http"

	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"golang.org/x/oauth2"
)

type (
	fakeAuthenticatorService struct {
		auth.AuthService
	}

	fakeOAuthClient struct {
		user *cloud.User
		oauthClient
	}

	fakeCloudClient struct {
		user *cloud.User
		cloud.Client
	}

	fakeUserSynchroniser struct{}
)

func (f *fakeAuthenticatorService) CreateSession(context.Context, auth.CreateSessionOptions) (*auth.Session, error) {
	return &auth.Session{}, nil
}

func (f *fakeOAuthClient) CallbackHandler(*http.Request) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func (f *fakeOAuthClient) NewClient(context.Context, *oauth2.Token) (cloud.Client, error) {
	return &fakeCloudClient{user: f.user}, nil
}

func (f *fakeCloudClient) GetUser(context.Context) (*cloud.User, error) {
	return f.user, nil
}

func (f *fakeUserSynchroniser) Sync(ctx context.Context, from cloud.User) error {
	return nil
}
