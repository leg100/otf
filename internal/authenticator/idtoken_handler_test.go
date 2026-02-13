package authenticator

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/leg100/otf/internal/user"
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

// Test_idtokenHandler_parseUserInfo tests extracting user info from an ID
// token.
func Test_idtokenHandler_parseUserInfo(t *testing.T) {

	// setup id token verifier
	fakeVerifier := func(key *rsa.PrivateKey) *oidc.IDTokenVerifier {
		keySet := &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{key.Public()}}
		return oidc.NewVerifier("", keySet, &oidc.Config{ClientID: "otf"})
	}

	// test extracting username from different claims
	tests := []struct {
		claim claim
		want  UserInfo
	}{
		{
			NameClaim,
			UserInfo{
				Username:  user.MustUsername("bobby"),
				AvatarURL: new("https://mypic.com"),
			},
		},
		{
			EmailClaim,
			UserInfo{
				Username: user.MustUsername("foo@example.com"),
			},
		},
		{
			SubClaim,
			UserInfo{
				Username: user.MustUsername("1111112222"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(string(tt.claim), func(t *testing.T) {
			// construct fake id token with wanted claims
			builder := jwt.NewBuilder().
				Audience([]string{"otf"}).
				Claim(string(tt.claim), tt.want.Username.String()).
				IssuedAt(time.Now()).
				Expiration(time.Now().Add(time.Minute))
			if tt.want.AvatarURL != nil {
				builder = builder.Claim("picture", tt.want.AvatarURL)
			}
			token, err := builder.Build()
			require.NoError(t, err)

			key, err := rsa.GenerateKey(rand.Reader, 1024)
			require.NoError(t, err)

			signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, key))
			require.NoError(t, err)

			handler := idTokenHandler{
				verifier:      fakeVerifier(key),
				usernameClaim: tt.claim,
			}
			got, err := handler.parseUserInfo(context.Background(), (&oauth2.Token{}).WithExtra(
				map[string]any{"id_token": string(signed)},
			))
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}
