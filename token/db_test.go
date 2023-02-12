package token

import (
	"context"
	"testing"

	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/user"
	"github.com/stretchr/testify/require"
)

func TestToken_CreateToken(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	tokenDB := newPGDB(db)

	user := user.CreateTestUser(t, db)
	token, err := NewToken(user.ID(), "testing")
	require.NoError(t, err)

	defer tokenDB.DeleteToken(ctx, token.ID())

	err = tokenDB.CreateToken(ctx, token)
	require.NoError(t, err)
}

func TestToken_DeleteToken(t *testing.T) {
	db := sql.NewTestDB(t)
	tokenDB := newPGDB(db)

	user := user.CreateTestUser(t, db)
	token := createTestToken(t, db, user.ID(), "testing")

	err := tokenDB.DeleteToken(context.Background(), token.ID())
	require.NoError(t, err)
}
