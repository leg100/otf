package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestToken(t *testing.T, db *pgdb, userID, description string) *Token {
	ctx := context.Background()

	token, err := NewToken(userID, description)
	require.NoError(t, err)

	err = db.CreateToken(ctx, token)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteToken(ctx, token.Token())
	})
	return token
}
