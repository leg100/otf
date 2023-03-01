package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, db *pgdb, opts ...NewUserOption) *User {
	ctx := context.Background()

	user := NewUser(uuid.NewString(), opts...)
	err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteUser(ctx, otf.UserSpec{UserID: &user.id})
	})
	return user
}
