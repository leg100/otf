package registry

import (
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func newTestSession(t *testing.T, org *otf.Organization, opts ...NewTestRegistrySessionOption) *Session {
	session, err := newSession(org.Name())
	require.NoError(t, err)

	for _, o := range opts {
		o(session)
	}

	return session
}

type NewTestRegistrySessionOption func(*Session)

func OverrideTestRegistrySessionExpiry(expiry time.Time) NewTestRegistrySessionOption {
	return func(session *Session) {
		session.expiry = expiry
	}
}
