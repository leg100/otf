package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func TestUser_CreateToken(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	token, err := otf.NewToken(user.ID, "testing")
	require.NoError(t, err)

	defer db.TokenStore().DeleteToken(context.Background(), token.ID)

	err = db.TokenStore().CreateToken(context.Background(), token)
	require.NoError(t, err)
}

func TestUser_DeleteToken(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	token := createTestToken(t, db, user.ID, "testing")

	err := db.TokenStore().DeleteToken(context.Background(), token.ID)
	require.NoError(t, err)
}
