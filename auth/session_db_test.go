package auth

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	ctx := context.Background()
	r := httptest.NewRequest("", "/", nil)
	db := newTestDB(t)
	user := createTestUser(t, db)

	t.Run("create", func(t *testing.T) {
		session, err := newSession(r, user.ID())
		require.NoError(t, err)

		err = db.createSession(ctx, session)
		require.NoError(t, err)

		db.deleteSession(ctx, session.Token())
	})

	t.Run("get by token", func(t *testing.T) {
		want := createTestSession(t, db, user.ID(), nil)

		got, err := db.getSessionByToken(ctx, want.Token())
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		session1 := createTestSession(t, db, user.ID(), nil)
		session2 := createTestSession(t, db, user.ID(), nil)

		// Retrieve all sessions
		sessions, err := db.listSessions(ctx, user.ID())
		require.NoError(t, err)

		assert.Contains(t, sessions, session1)
		assert.Contains(t, sessions, session2)
	})

	t.Run("purge expired sessions", func(t *testing.T) {
		_ = createTestSession(t, db, user.ID(), otf.Time(time.Now()))
		_ = createTestSession(t, db, user.ID(), otf.Time(time.Now()))

		_, err := db.DeleteSessionsExpired(ctx)
		require.NoError(t, err)

		sessions, err := db.listSessions(ctx, user.ID())
		require.NoError(t, err)
		assert.Equal(t, 0, len(sessions))
	})
}
