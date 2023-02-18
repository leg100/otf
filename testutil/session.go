package testutil

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/require"
)

func CreateSession(t *testing.T, db otf.DB, userID string, expiry *time.Time) *auth.Session {
	ctx := context.Background()
	svc := NewAuthService(t, db)

	session, err := svc.CreateSession(httptest.NewRequest("", "/", nil), userID)
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.DeleteSession(ctx, session.Token())
	})
	return session
}
