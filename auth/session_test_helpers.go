package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestSession(t *testing.T, username string, expiry *time.Time) *Session {
	r := httptest.NewRequest("", "/", nil)
	session, err := newSession(CreateSessionOptions{
		Request:  r,
		Username: &username,
		Expiry:   expiry,
	})
	require.NoError(t, err)

	return session
}
