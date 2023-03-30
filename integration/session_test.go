package integration

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		user := svc.createUser(t, ctx)
		_, err := svc.CreateSession(ctx, auth.CreateSessionOptions{
			Request:  httptest.NewRequest("", "/", nil),
			Username: &user.Username,
		})
		require.NoError(t, err)
	})

	t.Run("get by token", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createSession(t, ctx, nil, nil)

		got, err := svc.GetSession(ctx, want.Token())
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		user := svc.createUser(t, ctx)
		session1 := svc.createSession(t, ctx, user, nil)
		session2 := svc.createSession(t, ctx, user, nil)

		// Retrieve all sessions
		sessions, err := svc.ListSessions(ctx, user.Username)
		require.NoError(t, err)

		assert.Contains(t, sessions, session1)
		assert.Contains(t, sessions, session2)
	})

	t.Run("purge expired sessions", func(t *testing.T) {
		svc := setup(t, nil)
		session1 := svc.createSession(t, ctx, nil, nil)
		session2 := svc.createSession(t, ctx, nil, nil)

		ctx, cancel := context.WithCancel(ctx)
		svc.StartExpirer(ctx)
		defer cancel()

		_, err := svc.GetRegistrySession(ctx, session1.Token())
		assert.Equal(t, otf.ErrResourceNotFound, err)

		_, err = svc.GetRegistrySession(ctx, session2.Token())
		assert.Equal(t, otf.ErrResourceNotFound, err)
	})
}
