package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenDB(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	t.Run("create", func(t *testing.T) {

		user := createTestUser(t, db)
		token, err := NewToken(user.ID, "testing")
		require.NoError(t, err)

		defer db.DeleteToken(ctx, token.ID)

		err = db.CreateToken(ctx, token)
		require.NoError(t, err)
	})

	t.Run("delete", func(t *testing.T) {
		user := createTestUser(t, db)
		token := createTestToken(t, db, user.ID, "testing")

		err := db.DeleteToken(ctx, token.ID)
		require.NoError(t, err)
	})
}

func createTestToken(t *testing.T, db *pgdb, userID, description string) *Token {
	ctx := context.Background()

	token, err := NewToken(userID, description)
	require.NoError(t, err)

	err = db.CreateToken(ctx, token)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteToken(ctx, token.Token)
	})
	return token
}
