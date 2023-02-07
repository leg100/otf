package authenticator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestOAuthClient_RequestHandler(t *testing.T) {
	client := newTestOAuthServerClient(t)

	r := httptest.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()
	client.RequestHandler(w, r)

	assert.Equal(t, http.StatusFound, w.Result().StatusCode)

	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "https://otf-server.com/oauth/fake-cloud/callback", loc.Query().Get("redirect_uri"))

	if assert.Equal(t, 1, len(w.Result().Cookies())) {
		assert.Equal(t, w.Result().Cookies()[0].Value, loc.Query().Get("state"))
	}
}

func TestOAuthClient_CallbackHandler(t *testing.T) {
	client := newTestOAuthServerClient(t)
	r := httptest.NewRequest("GET", "/auth?state=state", nil)
	r.AddCookie(&http.Cookie{Name: oauthCookieName, Value: "state"})

	token, err := client.CallbackHandler(r)
	require.NoError(t, err)
	assert.Equal(t, token.AccessToken, "fake_token")
}

// newTestOAuthServerClient creates an OAuth server for testing purposes and
// returns a client configured to access the server.
func newTestOAuthServerClient(t *testing.T) *OAuthClient {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&oauth2.Token{AccessToken: "fake_token"})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	}))
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client, err := NewOAuthClient(NewOAuthClientConfig{
		CloudOAuthConfig: &cloud.CloudOAuthConfig{
			OAuthConfig: &oauth2.Config{
				Endpoint: oauth2.Endpoint{
					AuthURL:  srv.URL,
					TokenURL: srv.URL,
				},
			},
			Config: cloud.Config{
				SkipTLSVerification: true,
				Hostname:            u.Host,
				Name:                "fake-cloud",
			},
		},
		hostname: "otf-server.com",
	})
	require.NoError(t, err)
	return client
}
