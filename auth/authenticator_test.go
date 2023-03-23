package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewAuthenticators(t *testing.T) {
	opts := authenticatorOptions{
		Logger:          logr.Discard(),
		HostnameService: otf.FakeHostnameService{Host: "fake-host.org"},
		AuthService:     &fakeAuthenticatorService{},
		configs: []cloud.CloudOAuthConfig{
			{
				OAuthConfig: &oauth2.Config{
					ClientID:     "id-1",
					ClientSecret: "secret-1",
				},
			},
			{
				OAuthConfig: &oauth2.Config{
					ClientID:     "id-2",
					ClientSecret: "secret-2",
				},
			},
			{
				// should be skipped
				OAuthConfig: &oauth2.Config{},
			},
		},
	}
	got, err := newAuthenticators(opts)
	require.NoError(t, err)
	assert.Equal(t, 2, len(got))
}

func TestAuthenticator(t *testing.T) {
	cuser := cloud.User{
		Name: "fake-user",
		Teams: []cloud.Team{
			{
				Name:         "fake-team",
				Organization: "fake-org",
			},
		},
		Organizations: []string{"fake-org"},
	}

	t.Run("response_handler", func(t *testing.T) {
		authenticator := &authenticator{
			AuthService: &fakeAuthenticatorService{},
			oauthClient: &fakeOAuthClient{user: &cuser},
		}

		r := httptest.NewRequest("GET", "/auth?state=state", nil)
		r.AddCookie(&http.Cookie{Name: oauthCookieName, Value: "state"})
		w := httptest.NewRecorder()
		authenticator.responseHandler(w, r)

		assert.Equal(t, http.StatusFound, w.Result().StatusCode)

		loc, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/app/profile", loc.Path)

		if assert.Equal(t, 1, len(w.Result().Cookies())) {
			session := w.Result().Cookies()[0]
			assert.Equal(t, sessionCookie, session.Name)
		}
	})
}

type fakeAuthenticatorService struct {
	AuthService
}

func (f *fakeAuthenticatorService) sync(context.Context, cloud.User) (*User, error) {
	return NewUser("fake-user"), nil
}

func (f *fakeAuthenticatorService) CreateSession(context.Context, CreateSessionOptions) (*Session, error) {
	return &Session{}, nil
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
