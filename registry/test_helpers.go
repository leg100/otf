package registry

import (
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestSession(t *testing.T, org *otf.Organization, opts ...NewTestSessionOption) *Session {
	session, err := newSession(org.Name())
	require.NoError(t, err)

	for _, o := range opts {
		o(session)
	}

	return session
}

type NewTestSessionOption func(*Session)

func OverrideTestRegistrySessionExpiry(expiry time.Time) NewTestSessionOption {
	return func(session *Session) {
		session.expiry = expiry
	}
}
