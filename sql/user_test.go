package sql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Create(t *testing.T) {
	db := newTestDB(t)
	user := newTestUser()

	err := db.UserStore().Create(context.Background(), user)
	require.NoError(t, err)
}

func TestUser_Get(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	//_ = createTestSession(t, db)

	got, err := db.UserStore().Get(context.Background(), user.Username)
	require.NoError(t, err)

	assert.Equal(t, got, user)
}

func TestUser_GetWithSessions(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	_ = createTestSession(t, db)
	_ = createTestSession(t, db)

	got, err := db.UserStore().Get(context.Background(), user.Username)
	require.NoError(t, err)

	assert.Equal(t, got, user)
}
