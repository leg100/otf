package tokens

import (
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/require"
)

func newTestJWT(t *testing.T, secret []byte, id resource.ID, lifetime time.Duration, claims ...string) string {
	t.Helper()

	claimsMap := make(map[string]string, len(claims)/2)
	for i := 0; i < len(claims); i += 2 {
		claimsMap[claims[i]] = claims[i+1]
	}
	f := &factory{key: newTestJWK(t, secret)}
	token, err := f.NewToken(NewTokenOptions{
		ID:     id,
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
