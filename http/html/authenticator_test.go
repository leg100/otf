package html

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
)

func TestAuthenticator_RequestHandler(t *testing.T) {
	authenticator := &Authenticator{
		newFakeCloud("gitlab.com", nil),
		&fakeAuthenticatorApp{},
	}

	r := httptest.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()
	authenticator.requestHandler(w, r)

	assert.Equal(t, http.StatusFound, w.Result().StatusCode)

	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "gitlab.com", loc.Host)
	assert.Equal(t, "http://example.com/oauth/fake/callback", loc.Query().Get("redirect_uri"))

	if assert.Equal(t, 1, len(w.Result().Cookies())) {
		assert.Equal(t, w.Result().Cookies()[0].Value, loc.Query().Get("state"))
	}
}

func TestAuthenticator_ResponseHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	team := otf.NewTeam("fake-team", org)
	user := otf.NewUser("fake-user", otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))

	// IdP stub
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&oauth2.Token{AccessToken: "fake_token"})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	}))
	defer srv.Close()
	srvURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	authenticator := &Authenticator{
		newFakeCloud(srvURL.Host, user),
		&fakeAuthenticatorApp{},
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

	authenticator := &Authenticator{nil, &fakeAuthenticatorApp{}}
	user, err := authenticator.synchronise(context.Background(), &fakeDirectoryClient{user})
	require.NoError(t, err)

	assert.Equal(t, "fake-user", user.Username())

	if assert.Equal(t, 2, len(user.Organizations)) {
		assert.Equal(t, org.Name(), user.Organizations[0].Name())
		assert.Equal(t, "fake-user", user.Organizations[1].Name())
	}

	if assert.Equal(t, 2, len(user.Teams)) {
		assert.Equal(t, "fake-team", user.Teams[0].Name())
		assert.Equal(t, org.Name(), user.Teams[0].OrganizationName())

		assert.Equal(t, "owners", user.Teams[1].Name())
		assert.Equal(t, "fake-user", user.Teams[1].OrganizationName())
	}
}

type fakeCloud struct {
	cloudConfig
	user *otf.User
}

func newFakeCloud(hostname string, user *otf.User) *fakeCloud {
	return &fakeCloud{
		cloudConfig: cloudConfig{
			cloudName:           "fake",
			endpoint:            oauth2github.Endpoint,
			hostname:            hostname,
			skipTLSVerification: true,
			OAuthCredentials: &OAuthCredentials{
				clientID:     "abc-123",
				clientSecret: "xyz-789",
			},
		},
		user: user,
	}
}

func (f *fakeCloud) NewDirectoryClient(context.Context, DirectoryClientOptions) (DirectoryClient, error) {
	return &fakeDirectoryClient{f.user}, nil
}

func (f *fakeCloud) NewCloud() (Cloud, error) { return nil, nil }

type fakeDirectoryClient struct {
	user *otf.User
}

func (f *fakeDirectoryClient) GetUser(context.Context) (*otf.User, error) {
	return f.user, nil
}

type fakeAuthenticatorApp struct {
	otf.Application
}

func (f *fakeAuthenticatorApp) EnsureCreatedUser(context.Context, string) (*otf.User, error) {
	return otf.NewUser("fake-user"), nil
}

func (f *fakeAuthenticatorApp) CreateSession(context.Context, *otf.User, *otf.SessionData) (*otf.Session, error) {
	return &otf.Session{}, nil
}

func (f *fakeAuthenticatorApp) EnsureCreatedOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewOrganization(opts)
}

func (f *fakeAuthenticatorApp) SyncUserMemberships(ctx context.Context, user *otf.User, orgs []*otf.Organization, teams []*otf.Team) (*otf.User, error) {
	user.Organizations = orgs
	user.Teams = teams
	return user, nil
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
