package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Create(t *testing.T) {
	db := newTestDB(t)
	user := otf.NewTestUser()

	defer db.UserStore().Delete(context.Background(), otf.UserSpec{Username: &user.Username})

	err := db.UserStore().Create(context.Background(), user)
	require.NoError(t, err)
}

func TestUser_AddOrganizationMembership(t *testing.T) {
	db := newTestDB(t)

	org := createTestOrganization(t, db)
	user := createTestUser(t, db)

	err := db.UserStore().AddOrganizationMembership(context.Background(), user.ID(), org.ID())
	require.NoError(t, err)

	got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
	require.NoError(t, err)

	assert.Contains(t, got.Organizations, org)
}

func TestUser_RemoveOrganizationMembership(t *testing.T) {
	db := newTestDB(t)

	org := createTestOrganization(t, db)
	user := createTestUser(t, db, otf.WithOrganizationMemberships(org))

	err := db.UserStore().RemoveOrganizationMembership(context.Background(), user.ID(), org.ID())
	require.NoError(t, err)

	got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
	require.NoError(t, err)

	assert.NotContains(t, got.Organizations, org)
}

func TestUser_Update_CurrentOrganization(t *testing.T) {
	db := newTestDB(t)

	user := createTestUser(t, db)

	// set current org
	user.CurrentOrganization = otf.String("enron")

	err := db.UserStore().SetCurrentOrganization(context.Background(), user.ID(), *user.CurrentOrganization)
	require.NoError(t, err)

	got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
	require.NoError(t, err)

	assert.Equal(t, "enron", *got.CurrentOrganization)
}

func TestUser_Get(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	session := createTestSession(t, db, user.ID())
	token := createTestToken(t, db, user.ID(), "testing")

	tests := []struct {
		name string
		spec otf.UserSpec
	}{
		{
			name: "username",
			spec: otf.UserSpec{Username: &user.Username},
		},
		{
			name: "session token",
			spec: otf.UserSpec{SessionToken: &session.Token},
		},
		{
			name: "auth token ID",
			spec: otf.UserSpec{AuthenticationTokenID: otf.String(token.ID())},
		},
		{
			name: "auth token",
			spec: otf.UserSpec{AuthenticationToken: &token.Token},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.UserStore().Get(context.Background(), tt.spec)
			require.NoError(t, err)

			assert.Equal(t, got.ID(), user.ID())
		})
	}

}

func TestUser_Get_WithSessions(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	_ = createTestSession(t, db, user.ID())
	_ = createTestSession(t, db, user.ID())

	got, err := db.UserStore().Get(context.Background(), otf.UserSpec{Username: &user.Username})
	require.NoError(t, err)

	assert.Equal(t, 2, len(got.Sessions))

}

func TestUser_List(t *testing.T) {
	db := newTestDB(t)
	user1 := createTestUser(t, db)
	user2 := createTestUser(t, db)
	user3 := createTestUser(t, db)

	users, err := db.UserStore().List(context.Background())
	require.NoError(t, err)

	assert.Contains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)
}

func TestUser_Delete(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)

	err := db.UserStore().Delete(context.Background(), otf.UserSpec{Username: &user.Username})
	require.NoError(t, err)

	// Verify zero users after deletion
	users, err := db.UserStore().List(context.Background())
	require.NoError(t, err)
	assert.NotContains(t, users, user)
}
