package html

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
	got, err := newAuthenticators(logr.Discard(), &fakeAuthenticatorApp{}, []*otf.CloudOAuthConfig{
		{
			Config: &oauth2.Config{
				ClientID:     "id-1",
				ClientSecret: "secret-1",
			},
		},
		{
			Config: &oauth2.Config{
				ClientID:     "id-2",
				ClientSecret: "secret-2",
			},
		},
		{
			// should be skipped
			Config: &oauth2.Config{},
		},
	})
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
		authenticator := &Authenticator{
			Application: &fakeAuthenticatorApp{},
			oauthClient: &fakeOAuthClient{user: &cuser},
		}

		r := httptest.NewRequest("GET", "/auth?state=state", nil)
		r.AddCookie(&http.Cookie{Name: oauthCookieName, Value: "state"})
		w := httptest.NewRecorder()
		authenticator.responseHandler(w, r)

		assert.Equal(t, http.StatusFound, w.Result().StatusCode)

		loc, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/profile", loc.Path)

		if assert.Equal(t, 1, len(w.Result().Cookies())) {
			session := w.Result().Cookies()[0]
			assert.Equal(t, sessionCookie, session.Name)
		}
	})

	t.Run("synchronise", func(t *testing.T) {
		authenticator := &Authenticator{
			Application: &fakeAuthenticatorApp{},
		}

		user, err := authenticator.synchronise(context.Background(), &fakeCloudClient{user: &cuser})
		require.NoError(t, err)

		assert.Equal(t, "fake-user", user.Username())

		if assert.Equal(t, 2, len(user.Organizations())) {
			assert.Equal(t, "fake-org", user.Organizations()[0].Name())
			assert.Equal(t, "fake-user", user.Organizations()[1].Name())
		}

		if assert.Equal(t, 2, len(user.Teams())) {
			assert.Equal(t, "fake-team", user.Teams()[0].Name())
			assert.Equal(t, "fake-org", user.Teams()[0].Organization().Name())

			assert.Equal(t, "owners", user.Teams()[1].Name())
			assert.Equal(t, "fake-user", user.Teams()[1].Organization().Name())
		}
	})
}

type fakeAuthenticatorApp struct {
	otf.Application
}

func (f *fakeAuthenticatorApp) Hostname() string { return "fake-host.org" }

func (f *fakeAuthenticatorApp) EnsureCreatedUser(context.Context, string) (*otf.User, error) {
	return otf.NewUser("fake-user"), nil
}

func (f *fakeAuthenticatorApp) CreateSession(context.Context, string, string) (*otf.Session, error) {
	return &otf.Session{}, nil
}

func (f *fakeAuthenticatorApp) EnsureCreatedOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewOrganization(opts)
}

func (f *fakeAuthenticatorApp) SyncUserMemberships(ctx context.Context, user *otf.User, orgs []*otf.Organization, teams []*otf.Team) (*otf.User, error) {
	err := user.SyncMemberships(ctx, &fakeUserStore{}, orgs, teams)
	return user, err
}

func (f *fakeAuthenticatorApp) EnsureCreatedTeam(ctx context.Context, opts otf.CreateTeamOptions) (*otf.Team, error) {
	org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
		Name: otf.String(opts.Organization),
	})
	if err != nil {
		return nil, err
	}
	return otf.NewTeam(opts.Name, org), nil
}

type fakeUserStore struct {
	otf.UserStore
}

func (f *fakeUserStore) AddOrganizationMembership(ctx context.Context, id, orgID string) error {
	return nil
}

func (f *fakeUserStore) RemoveOrganizationMembership(ctx context.Context, id, orgID string) error {
	return nil
}
func (f *fakeUserStore) AddTeamMembership(ctx context.Context, id, teamID string) error { return nil }
func (f *fakeUserStore) RemoveTeamMembership(ctx context.Context, id, teamID string) error {
	return nil
}

type fakeOAuthClient struct {
	user *cloud.User
	oauthClient
}

func (f *fakeOAuthClient) CallbackHandler(*http.Request) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func (f *fakeOAuthClient) NewClient(context.Context, *oauth2.Token) (otf.CloudClient, error) {
	return &fakeCloudClient{user: f.user}, nil
}

type fakeCloudClient struct {
	user *cloud.User
	otf.CloudClient
}

func (f *fakeCloudClient) GetUser(context.Context) (*cloud.User, error) {
	return f.user, nil
}
