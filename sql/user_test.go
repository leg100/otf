package sql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	user := newTestUser(org)

	err := db.UserStore().Create(context.Background(), user)
	require.NoError(t, err)
}

func TestUser_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	user := createTestUser(t, db, org)

	got, err := db.UserStore().Get(context.Background(), user.Username)
	require.NoError(t, err)

	assert.Equal(t, got, user)
}

func TestUser_Get_WithSessions(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	user := createTestUser(t, db, org)
	_ = createTestSession(t, db, user.ID)
	_ = createTestSession(t, db, user.ID)

	got, err := db.UserStore().Get(context.Background(), user.Username)
	require.NoError(t, err)

	assert.Equal(t, 2, len(got.Sessions))
}

func TestUser_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	_ = createTestUser(t, db, org)
	_ = createTestUser(t, db, org)
	_ = createTestUser(t, db, org)

	users, err := db.UserStore().List(context.Background(), org.ID)
	require.NoError(t, err)

	assert.Equal(t, 3, len(users))
}

func TestUser_Delete(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	user := createTestUser(t, db, org)

	err := db.UserStore().Delete(context.Background(), user.ID)
	require.NoError(t, err)

	// Verify zero users after deletion
	users, err := db.UserStore().List(context.Background(), org.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, len(users))
}

func TestUser_LinkSession(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	user := createTestUser(t, db, org)
	session := createTestSession(t, db, user.ID)

	err := db.UserStore().LinkSession(context.Background(), session.Token, user.ID)
	require.NoError(t, err)

	// Verify user has a session after linking
	user, err = db.UserStore().Get(context.Background(), user.Username)
	require.NoError(t, err)
	assert.Equal(t, 1, len(user.Sessions))
}
