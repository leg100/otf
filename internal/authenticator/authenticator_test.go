package authenticator

import (
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewAuthenticatorService(t *testing.T) {
	opts := Options{
		Logger:          logr.Discard(),
		HostnameService: internal.FakeHostnameService{Host: "fake-host.org"},
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

// TestLoginHandler tests the login page handler, testing for the presence of a
// login button for each configured cloud.
func TestLoginHandler(t *testing.T) {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)
	svc := &service{
		renderer: renderer,
	}

	svc.authenticators = []authenticator{
		&oauthAuthenticator{
			oauthClient: &OAuthClient{
				cloudConfig: cloud.Config{Name: "cloud1"},
			},
		},
		&oauthAuthenticator{
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
