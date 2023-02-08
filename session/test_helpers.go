package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createTestSession(t *testing.T, db otf.DB, userID string, opts ...otf.NewSessionOption) *otf.Session {
	session := NewTestSession(t, userID, opts...)
	ctx := context.Background()

	err := db.CreateSession(ctx, session)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteSession(ctx, session.Token())
	})
	return session
}

func NewTestSession(t *testing.T, userID string, opts ...NewSessionOption) *Session {
	session, err := NewSession(userID, "127.0.0.1")
	require.NoError(t, err)

	for _, o := range opts {
		o(session)
	}

	return session
}

type newTestDBOption func(*Options)

func overrideCleanupInterval(d time.Duration) newTestDBOption {
	return func(o *Options) {
		o.CleanupInterval = d
	}
}
