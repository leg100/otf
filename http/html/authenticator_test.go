package html

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestAuthenticator_ResponseHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	team := otf.NewTeam("fake-team", org)
	user := otf.NewUser("fake-user", otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))

	authenticator := &Authenticator{
		Application: &fakeAuthenticatorApp{},
		oauthClient: &fakeOAuthClient{user: user},
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
}

func TestAuthenticator_Synchronise(t *testing.T) {
	org := otf.NewTestOrganization(t)
	team := otf.NewTeam("fake-team", org)
	user := otf.NewUser("fake-user", otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))

	authenticator := &Authenticator{
		Application: &fakeAuthenticatorApp{},
		oauthClient: &fakeOAuthClient{user: user},
	}

	user, err := authenticator.synchronise(context.Background(), &fakeCloudClient{user: user})
	require.NoError(t, err)

	assert.Equal(t, "fake-user", user.Username())

	if assert.Equal(t, 2, len(user.Organizations())) {
		assert.Equal(t, org.Name(), user.Organizations()[0].Name())
		assert.Equal(t, "fake-user", user.Organizations()[1].Name())
	}

	if assert.Equal(t, 2, len(user.Teams())) {
		assert.Equal(t, "fake-team", user.Teams()[0].Name())
		assert.Equal(t, org.Name(), user.Teams()[0].Organization().Name())

		assert.Equal(t, "owners", user.Teams()[1].Name())
		assert.Equal(t, "fake-user", user.Teams()[1].Organization().Name())
	}
}

type fakeAuthenticatorApp struct {
	otf.Application
}

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

func (f *fakeAuthenticatorApp) EnsureCreatedTeam(ctx context.Context, name, organizationName string) (*otf.Team, error) {
	org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
		Name: otf.String(organizationName),
	})
	if err != nil {
		return nil, err
	}
	return otf.NewTeam(name, org), nil
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
	user *otf.User
	oauthClient
}

func (f *fakeOAuthClient) CallbackHandler(*http.Request) (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func (f *fakeOAuthClient) NewClient(context.Context, *oauth2.Token) (otf.CloudClient, error) {
	return &fakeCloudClient{user: f.user}, nil
}

type fakeCloudClient struct {
	user *otf.User
	otf.CloudClient
}

func (f *fakeCloudClient) GetUser(context.Context) (*otf.User, error) {
	return f.user, nil
}
