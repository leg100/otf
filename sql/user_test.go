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
	user := otf.NewUser("mr-t")

	defer db.DeleteUser(context.Background(), otf.UserSpec{Username: otf.String(user.Username())})

	err := db.CreateUser(context.Background(), user)
	require.NoError(t, err)
}

func TestUser_AddOrganizationMembership(t *testing.T) {
	db := newTestDB(t)

	org := createTestOrganization(t, db)
	user := createTestUser(t, db)

	err := db.AddOrganizationMembership(context.Background(), user.ID(), org.ID())
	require.NoError(t, err)

	got, err := db.GetUser(context.Background(), otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Contains(t, got.Organizations, org)
}

func TestUser_RemoveOrganizationMembership(t *testing.T) {
	db := newTestDB(t)

	org := createTestOrganization(t, db)
	user := createTestUser(t, db, otf.WithOrganizationMemberships(org))

	err := db.RemoveOrganizationMembership(context.Background(), user.ID(), org.ID())
	require.NoError(t, err)

	got, err := db.GetUser(context.Background(), otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.NotContains(t, got.Organizations, org)
}

func TestUser_Get(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	session := createTestSession(t, db, user.ID())
	// ...and token
	token := createTestToken(t, db, user.ID(), "testing")

	tests := []struct {
		name string
		spec otf.UserSpec
	}{
		{
			name: "username",
			spec: otf.UserSpec{Username: otf.String(user.Username())},
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
			spec: otf.UserSpec{AuthenticationToken: otf.String(token.Token())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetUser(context.Background(), tt.spec)
			require.NoError(t, err)

			assert.Equal(t, got.ID(), user.ID())
		})
	}

}

func TestUser_Get_NotFound(t *testing.T) {
	db := newTestDB(t)

	_, err := db.GetUser(context.Background(), otf.UserSpec{Username: otf.String("does-not-exist")})
	assert.Equal(t, otf.ErrResourceNotFound, err)
}

func TestUser_Get_WithSessions(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)
	_ = createTestSession(t, db, user.ID())
	_ = createTestSession(t, db, user.ID())

	got, err := db.GetUser(context.Background(), otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Equal(t, 2, len(got.Sessions))
}

func TestUser_List(t *testing.T) {
	db := newTestDB(t)
	user1 := createTestUser(t, db)
	user2 := createTestUser(t, db)
	user3 := createTestUser(t, db)

	users, err := db.ListUsers(context.Background())
	require.NoError(t, err)

	assert.Contains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)
}

func TestUser_Delete(t *testing.T) {
	db := newTestDB(t)
	user := createTestUser(t, db)

	err := db.DeleteUser(context.Background(), otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	// Verify zero users after deletion
	users, err := db.ListUsers(context.Background())
	require.NoError(t, err)
	assert.NotContains(t, users, user)
}
