package authenticator

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewAuthenticators(t *testing.T) {
	opts := Options{
		Logger:          logr.Discard(),
		HostnameService: otf.FakeHostnameService{Host: "fake-host.org"},
		AuthService:     &fakeAuthenticatorService{},
		Configs: []cloud.CloudOAuthConfig{
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
	got, err := NewAuthenticatorService(opts)
	require.NoError(t, err)
	assert.Equal(t, 2, len(got.authenticators))
}

func TestAuthenticator_ResponseHandler(t *testing.T) {
	cuser := cloud.User{
		Name: "fake-user",
		Teams: []cloud.Team{
			{
				Name:         "fake-team",
				Organization: "fake-org",
			},
		},
	}

	authenticator := &authenticator{
		AuthService:      &fakeAuthenticatorService{},
		oauthClient:      &fakeOAuthClient{user: &cuser},
		userSynchroniser: &fakeUserSynchroniser{},
	}

	r := httptest.NewRequest("GET", "/auth?state=state", nil)
	r.AddCookie(&http.Cookie{Name: oauthCookieName, Value: "state"})
	w := httptest.NewRecorder()
	authenticator.responseHandler(w, r)

	assert.Equal(t, http.StatusFound, w.Result().StatusCode)

	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "/app/profile", loc.Path)

	assert.Equal(t, 1, len(w.Result().Cookies()))
}

// TestLoginHandler tests the login page handler, testing for the presence of a
// login button for each configured cloud.
func TestLoginHandler(t *testing.T) {
	renderer, err := html.NewViewEngine(false)
	require.NoError(t, err)
	svc := &service{
		Renderer: renderer,
	}

	svc.authenticators = []*authenticator{
		{
			oauthClient: &OAuthClient{
				cloudConfig: cloud.Config{Name: "cloud1"},
			},
		},
		{
			oauthClient: &OAuthClient{
				cloudConfig: cloud.Config{Name: "cloud2"},
			},
		},
	}

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	svc.loginHandler(w, r)
	body := w.Body.String()
	if assert.Equal(t, 200, w.Code, "output: %s", body) {
		assert.Contains(t, body, "Login with Cloud1")
		assert.Contains(t, body, "Login with Cloud2")
	}
}
