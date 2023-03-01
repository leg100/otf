package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/require"
)

func CreateUser(t *testing.T, db otf.DB, opts ...auth.NewUserOption) *auth.User {
	ctx := context.Background()
	svc := NewAuthService(t, db)

	user, err := svc.CreateUser(ctx, uuid.NewString(), opts...)
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.DeleteUser(ctx, user.Username())
	})
	return user
}