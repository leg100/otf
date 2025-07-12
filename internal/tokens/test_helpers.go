package tokens

import (
	"testing"
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/require"
)

func newTestJWT(t *testing.T, secret []byte, id resource.TfeID, lifetime time.Duration) string {
	t.Helper()

	f := &tokenFactory{symKey: newTestJWK(t, secret)}
	token, err := f.NewToken(id, WithExpiry(time.Now().Add(lifetime)))
	require.NoError(t, err)
	return string(token)
}

func newTestJWK(t *testing.T, secret []byte) jwk.Key {
	t.Helper()

	key, err := jwk.FromRaw(secret)
	require.NoError(t, err)
	return key
}
