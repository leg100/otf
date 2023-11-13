package tokens

import (
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/require"
)

func NewTestSessionJWT(t *testing.T, username string, secret []byte, lifetime time.Duration) string {
	t.Helper()

	return newTestJWT(t, secret, userSessionKind, lifetime, "sub", username)
}

func newTestJWT(t *testing.T, secret []byte, kind Kind, lifetime time.Duration, claims ...string) string {
	t.Helper()

	claimsMap := make(map[string]string, len(claims)/2)
	for i := 0; i < len(claims); i += 2 {
		claimsMap[claims[i]] = claims[i+1]
	}
	f := &factory{key: newTestJWK(t, secret)}
	token, err := f.NewToken(NewTokenOptions{
		Kind:   kind,
		Expiry: internal.Time(time.Now().Add(lifetime)),
		Claims: claimsMap,
	})
	require.NoError(t, err)
	return string(token)
}

func newTestJWK(t *testing.T, secret []byte) jwk.Key {
	t.Helper()

	key, err := jwk.FromRaw(secret)
	require.NoError(t, err)
	return key
}
