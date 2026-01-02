package authenticator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestOAuthClient_requestHandler(t *testing.T) {
	client := newTestOAuthServerClient(t, testutils.ParseID(t, "user-bobby"))

	r := httptest.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()
	client.requestHandler(w, r)

	assert.Equal(t, http.StatusFound, w.Result().StatusCode)

	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "https://otf-server.com/oauth/fake-cloud/callback", loc.Query().Get("redirect_uri"))

	if assert.Equal(t, 1, len(w.Result().Cookies())) {
		assert.Equal(t, w.Result().Cookies()[0].Value, loc.Query().Get("state"))
	}
}

func TestOAuthClient_callbackHandler(t *testing.T) {
	client := newTestOAuthServerClient(t, testutils.ParseID(t, "user-bobby"))
	r := httptest.NewRequest("GET", "/auth?state=state", nil)
	r.AddCookie(&http.Cookie{Name: oauthCookieName, Value: "state"})
	w := httptest.NewRecorder()

	client.callbackHandler(w, r)
	assert.Equal(t, "user-bobby", w.Header().Get("user-id"))
}

// newTestOAuthServerClient creates an OAuth server for testing purposes and
// returns a client configured to access the server.
func newTestOAuthServerClient(t *testing.T, userID resource.TfeID) *OAuthClient {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&oauth2.Token{AccessToken: "fake_token"})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	}))
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client, err := newOAuthClient(
		logr.Discard(),
		fakeTokenHandler{},
		internal.NewHostnameService("otf-server.com"),
		&fakeTokensService{},
		&fakeUserService{userID},
		OAuthConfig{
			BaseURL: &internal.WebURL{URL: *u},
			Endpoint: oauth2.Endpoint{
				AuthURL:  srv.URL,
				TokenURL: srv.URL,
			},
			Name:                "fake-cloud",
			SkipTLSVerification: true,
		},
	)
	require.NoError(t, err)
	return client
}
