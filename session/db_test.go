package session

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_CreateSession(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	sessionDB := &DB{db}

	user := user.CreateTestUser(t, db)
	session := otf.NewTestSession(t, user.ID())

	defer sessionDB.DeleteSession(ctx, session.Token())

	err := sessionDB.CreateSession(ctx, session)
	require.NoError(t, err)
}

func TestSession_GetByToken(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	sessionDB := &DB{db}

	user := user.CreateTestUser(t, db)
	want := createTestSession(t, db, user.ID())

	got, err := sessionDB.GetSessionByToken(ctx, want.Token())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestSession_List(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	sessionDB := &DB{db}

	user := user.CreateTestUser(t, db)
	session1 := createTestSession(t, db, user.ID())
	session2 := createTestSession(t, db, user.ID())

	// Retrieve all sessions
	sessions, err := sessionDB.ListSessions(ctx, user.ID())
	require.NoError(t, err)

	assert.Contains(t, sessions, session1)
	assert.Contains(t, sessions, session2)
}

func TestSession_SessionCleanup(t *testing.T) {
	ctx := context.Background()

	db := sql.NewTestDB(t)
	sessionDB := &DB{db}

	user := user.CreateTestUser(t, db)

	_ = createTestSession(t, db, user.ID(), SessionExpiry(time.Now()))
	_ = createTestSession(t, db, user.ID(), SessionExpiry(time.Now()))

	err := sessionDB.deleteExpired()
	require.NoError(t, err)

	sessions, err := sessionDB.ListSessions(ctx, user.ID())
	require.NoError(t, err)
	assert.Equal(t, 0, len(sessions))
}
