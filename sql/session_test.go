package sql

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_CreateSession(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	session := newTestSession(t, user.ID())

	defer db.DeleteSession(context.Background(), session.Token)

	err := db.CreateSession(context.Background(), session)
	require.NoError(t, err)
}

// TestSession_SessionCleanup tests the session cleanup background routine. We
// override the cleanup interval to just every 100ms, so after waiting for 300ms
// the sessions should be cleaned up.
func TestSession_SessionCleanup(t *testing.T) {
	db := newTestDB(t, 100*time.Millisecond)
	user := createTestUser(t, db)

	_ = createTestSession(t, db, user.ID(), overrideExpiry(time.Now()))
	_ = createTestSession(t, db, user.ID(), overrideExpiry(time.Now()))

	time.Sleep(300 * time.Millisecond)

	got, err := db.GetUser(context.Background(), otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Equal(t, 0, len(got.Sessions))
}
