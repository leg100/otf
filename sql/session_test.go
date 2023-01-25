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
	db := NewTestDB(t)
	user := createTestUser(t, db)
	session := otf.NewTestSession(t, user.ID())

	defer db.DeleteSession(context.Background(), session.Token())

	err := db.CreateSession(context.Background(), session)
	require.NoError(t, err)
}

func TestSession_GetByToken(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	user := createTestUser(t, db)
	want := createTestSession(t, db, user.ID())

	got, err := db.GetSessionByToken(ctx, want.Token())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestSession_List(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	user := createTestUser(t, db)
	session1 := createTestSession(t, db, user.ID())
	session2 := createTestSession(t, db, user.ID())

	// Retrieve all sessions
	sessions, err := db.ListSessions(ctx, user.ID())
	require.NoError(t, err)

	assert.Contains(t, sessions, session1)
	assert.Contains(t, sessions, session2)
}

// TestSession_SessionCleanup tests the session cleanup background routine. We
// override the cleanup interval to just every 100ms, so after waiting for 300ms
// the sessions should be cleaned up.
func TestSession_SessionCleanup(t *testing.T) {
	ctx := context.Background()

	db := NewTestDB(t, overrideCleanupInterval(100*time.Millisecond))
	user := createTestUser(t, db)

	_ = createTestSession(t, db, user.ID(), otf.SessionExpiry(time.Now()))
	_ = createTestSession(t, db, user.ID(), otf.SessionExpiry(time.Now()))

	time.Sleep(300 * time.Millisecond)

	sessions, err := db.ListSessions(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, 0, len(sessions))
}
