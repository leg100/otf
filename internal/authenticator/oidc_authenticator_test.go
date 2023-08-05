package authenticator

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOIDCAuthenticator(t *testing.T) {
	ctx := context.Background()
	_, err := newOIDCAuthenticator(ctx, oidcAuthenticatorOptions{})
	assert.Equal(t, ErrMissingOIDCIssuerURL, err)
}

func TestOIDCAuthenticator_ResponseHandler(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	authenticator := &oidcAuthenticator{
		TokensService: &fakeAuthenticatorService{},
		oauthClient: &fakeOAuthClient{
			token: fakeOAuthToken(t, "", "otf", priv),
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
