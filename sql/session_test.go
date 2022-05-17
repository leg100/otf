package sql

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_CreateSession(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	session := newTestSession(t, user.ID)

	defer db.SessionStore().DeleteSession(context.Background(), session.Token)

	err := db.SessionStore().CreateSession(context.Background(), session)
	require.NoError(t, err)
}

func TestUser_Flash(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	session := createTestSession(t, db, user.ID)

	flash := &otf.Flash{
		Type:    otf.FlashSuccessType,
		Message: "test succeeded",
	}

	err := db.SessionStore().SetFlash(context.Background(), session.Token, flash)
	require.NoError(t, err)

	got, err := db.SessionStore().PopFlash(context.Background(), session.Token)
	require.NoError(t, err)

	assert.Equal(t, flash, got)
}

// TestUser_SessionCleanup tests the session cleanup background routine. We
// override the cleanup interval to just every 100ms, so after waiting for 300ms
// the sessions should be cleaned up.
func TestUser_SessionCleanup(t *testing.T) {
	db := newTestDB(t, 100*time.Millisecond)
	user := createTestUser(t, db)

	_ = createTestSession(t, db, user.ID, overrideExpiry(time.Now()))
	_ = createTestSession(t, db, user.ID, overrideExpiry(time.Now()))

	time.Sleep(300 * time.Millisecond)

	got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
	require.NoError(t, err)

	assert.Equal(t, 0, len(got.Sessions))
}
