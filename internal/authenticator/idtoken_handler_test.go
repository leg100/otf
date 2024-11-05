package authenticator

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/leg100/otf/internal/resource"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func Test_newIDTokenHandler(t *testing.T) {
	ctx := context.Background()
	_, err := newIDTokenHandler(ctx, OIDCConfig{})
	assert.Equal(t, ErrMissingOIDCIssuerURL, err)
}

// Test_idtokenHandler_getUsername tests extracting the 'name' claim from an ID
// token.
func Test_idtokenHandler_getUsername(t *testing.T) {
	// create id token
	token, err := jwt.NewBuilder().
		Audience([]string{"otf"}).
		Claim("name", "bobby").
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(time.Minute)).
		Build()
	require.NoError(t, err)
	key, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, key))
	require.NoError(t, err)

	// setup id token verifier
	fakeVerifier := func(t *testing.T, aud string, key *rsa.PrivateKey) *oidc.IDTokenVerifier {
		keySet := &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{key.Public()}}
		return oidc.NewVerifier("", keySet, &oidc.Config{ClientID: resource.ParseID("otf")})
	}
	// setup handler to parse the 'name' claim
	username, err := newUsernameClaim("name")
	require.NoError(t, err)

	handler := idtokenHandler{
		verifier: fakeVerifier(t, "otf", key),
		username: username,
	}
	got, err := handler.getUsername(context.Background(), (&oauth2.Token{}).WithExtra(
		map[string]any{"id_token": string(signed)},
	))
	require.NoError(t, err)
	assert.Equal(t, "bobby", got)
}
