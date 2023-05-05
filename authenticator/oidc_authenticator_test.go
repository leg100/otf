package authenticator

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCAuthenticator_ResponseHandler(t *testing.T) {
	user := cloud.User{
		Name: "fake-user",
		Teams: []cloud.Team{
			{
				Name:         "fake-team",
				Organization: "fake-org",
			},
		},
	}
	priv, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	authenticator := &oidcAuthenticator{
		TokensService: &fakeAuthenticatorService{},
		oauthClient: &fakeOAuthClient{
			user:  &user,
			token: fakeIDToken(t, "otf", priv),
		},
		verifier: fakeVerifier(t, "otf", priv),
	}

	r := httptest.NewRequest("GET", "/auth?state=state", nil)
	r.AddCookie(&http.Cookie{Name: oauthCookieName, Value: "state"})
	w := httptest.NewRecorder()
	authenticator.ResponseHandler(w, r)

	assert.Equal(t, http.StatusFound, w.Result().StatusCode, w.Body.String())

	loc, err := w.Result().Location()
	require.NoError(t, err)
	assert.Equal(t, "/app/profile", loc.Path)
}
