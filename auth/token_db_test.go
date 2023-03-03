package auth

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func createTestToken(t *testing.T, db *pgdb, userID, description string) *otf.Token {
	ctx := context.Background()

	token, err := otf.NewToken(userID, description)
	require.NoError(t, err)

	err = db.CreateToken(ctx, token)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteToken(ctx, token.Token)
	})
	return token
}
