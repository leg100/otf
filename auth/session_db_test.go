package auth

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_CreateSession(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	user := CreateTestUser(t, db)
	session := newTestSession(t, user.ID(), nil)

	defer db.deleteSession(ctx, session.Token())

	err := db.createSession(ctx, session)
	require.NoError(t, err)
}

func TestSession_GetByToken(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	user := CreateTestUser(t, db)
	want := createTestSession(t, db, user.ID(), nil)

	got, err := db.getSessionByToken(ctx, want.Token())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestSession_List(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	user := CreateTestUser(t, db)
	session1 := createTestSession(t, db, user.ID(), nil)
	session2 := createTestSession(t, db, user.ID(), nil)

	// Retrieve all sessions
	sessions, err := db.listSessions(ctx, user.ID())
	require.NoError(t, err)

	assert.Contains(t, sessions, session1)
	assert.Contains(t, sessions, session2)
}

func TestSession_SessionCleanup(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	user := createTestUser(t, db)

	_ = createTestSession(t, db, user.ID(), otf.Time(time.Now()))
	_ = createTestSession(t, db, user.ID(), otf.Time(time.Now()))

	err := db.deleteExpired(ctx)
	require.NoError(t, err)

	sessions, err := db.listSessions(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, 0, len(sessions))
}
