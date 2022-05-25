package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func TestToken_CreateToken(t *testing.T) {
	ctx := context.Background()

	db := newTestDB(t)
	user := createTestUser(t, db)
	token, err := otf.NewToken(user.ID(), "testing")
	require.NoError(t, err)

	defer db.TokenStore().DeleteToken(ctx, token.ID())

	err = db.TokenStore().CreateToken(ctx, token)
	require.NoError(t, err)
}

func TestToken_DeleteToken(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	token := createTestToken(t, db, user.ID(), "testing")

	err := db.TokenStore().DeleteToken(context.Background(), token.ID())
	require.NoError(t, err)
}
