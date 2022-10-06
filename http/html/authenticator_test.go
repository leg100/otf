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
		newFakeCloud("gitlab.com"),
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
		newFakeCloud(srvURL.Host),
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

type fakeCloud struct {
	cloudConfig
}

func newFakeCloud(hostname string) *fakeCloud {
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
	}
}

func (f *fakeCloud) NewDirectoryClient(context.Context, DirectoryClientOptions) (DirectoryClient, error) {
	return &fakeDirectoryClient{}, nil
}

func (f *fakeCloud) NewCloud() (Cloud, error) { return nil, nil }

type fakeDirectoryClient struct{}

func (f *fakeDirectoryClient) GetUser(context.Context) (string, error) {
	return "fake-user", nil
}

func (f *fakeDirectoryClient) ListOrganizations(context.Context) ([]string, error) {
	return []string{"fake-org"}, nil
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

func (f *fakeAuthenticatorApp) SyncOrganizationMemberships(ctx context.Context, user *otf.User, orgs []*otf.Organization) (*otf.User, error) {
	return user, nil
}
