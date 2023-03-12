package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createTestRegistrySession(t *testing.T, db *pgdb, org string, expiry *time.Time) *RegistrySession {
	ctx := context.Background()

	session, err := NewRegistrySession(org)
	require.NoError(t, err)

	// optionally override expiry
	if expiry != nil {
		session.Expiry = *expiry
	}

	err = db.createRegistrySession(ctx, session)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.deleteSession(ctx, session.Token)
	})
	return session
}
