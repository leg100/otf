package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Create(t *testing.T) {
	db := NewTestDB(t)
	user := otf.NewUser("mr-t")

	defer db.DeleteUser(context.Background(), otf.UserSpec{Username: otf.String(user.Username())})

	err := db.CreateUser(context.Background(), user)
	require.NoError(t, err)
}

func TestUser_AddOrganizationMembership(t *testing.T) {
	db := NewTestDB(t)
	ctx := context.Background()

	org := CreateTestOrganization(t, db)
	user := CreateTestUser(t, db)

	err := db.AddOrganizationMembership(ctx, user.ID(), org.Name())
	require.NoError(t, err)

	got, err := db.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Contains(t, got.Organizations(), org.Name())
}

func TestUser_RemoveOrganizationMembership(t *testing.T) {
	db := NewTestDB(t)
	ctx := context.Background()

	org := CreateTestOrganization(t, db)
	user := CreateTestUser(t, db, otf.WithOrganizationMemberships(org.Name()))

	err := db.RemoveOrganizationMembership(ctx, user.ID(), org.Name())
	require.NoError(t, err)

	got, err := db.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.NotContains(t, got.Organizations(), org)
}

func TestUser_AddTeamMembership(t *testing.T) {
	db := NewTestDB(t)
	ctx := context.Background()

	org := CreateTestOrganization(t, db)
	team := createTestTeam(t, db, org)
	user := CreateTestUser(t, db, otf.WithOrganizationMemberships(org.Name()))

	err := db.AddTeamMembership(ctx, user.ID(), team.ID())
	require.NoError(t, err)

	got, err := db.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Contains(t, got.Teams(), team)
}

func TestUser_RemoveTeamMembership(t *testing.T) {
	db := NewTestDB(t)
	ctx := context.Background()

	org := CreateTestOrganization(t, db)
	team := createTestTeam(t, db, org)
	user := CreateTestUser(t, db, otf.WithOrganizationMemberships(org.Name()), otf.WithTeamMemberships(team))

	err := db.RemoveTeamMembership(ctx, user.ID(), team.ID())
	require.NoError(t, err)

	got, err := db.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.NotContains(t, got.Teams(), team)
}

func TestUser_Get(t *testing.T) {
	db := NewTestDB(t)

	org1 := CreateTestOrganization(t, db)
	org2 := CreateTestOrganization(t, db)
	team1 := createTestTeam(t, db, org1)
	team2 := createTestTeam(t, db, org2)

	user := CreateTestUser(t, db,
		otf.WithOrganizationMemberships(org1.Name(), org2.Name()),
		otf.WithTeamMemberships(team1, team2))

	session1 := createTestSession(t, db, user.ID())
	_ = createTestSession(t, db, user.ID())

	token1 := createTestToken(t, db, user.ID(), "testing")
	_ = createTestToken(t, db, user.ID(), "testing")

	tests := []struct {
		name string
		spec otf.UserSpec
	}{
		{
			name: "id",
			spec: otf.UserSpec{UserID: otf.String(user.ID())},
		},
		{
			name: "username",
			spec: otf.UserSpec{Username: otf.String(user.Username())},
		},
		{
			name: "session token",
			spec: otf.UserSpec{SessionToken: otf.String(session1.Token())},
		},
		{
			name: "auth token",
			spec: otf.UserSpec{AuthenticationToken: otf.String(token1.Token())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetUser(context.Background(), tt.spec)
			require.NoError(t, err)

			assert.Equal(t, got.ID(), user.ID())
			assert.Equal(t, got.Username(), user.Username())
			assert.Equal(t, got.CreatedAt(), user.CreatedAt())
			assert.Equal(t, got.UpdatedAt(), user.UpdatedAt())
			assert.Equal(t, 2, len(got.Organizations()))
			assert.Equal(t, 2, len(got.Teams()))
		})
	}
}

func TestUser_Get_NotFound(t *testing.T) {
	db := NewTestDB(t)

	_, err := db.GetUser(context.Background(), otf.UserSpec{Username: otf.String("does-not-exist")})
	assert.Equal(t, otf.ErrResourceNotFound, err)
}

func TestUser_List(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	team := createTestTeam(t, db, org)
	user1 := CreateTestUser(t, db)
	user2 := CreateTestUser(t, db, otf.WithOrganizationMemberships(org.Name()))
	user3 := CreateTestUser(t, db, otf.WithOrganizationMemberships(org.Name()), otf.WithTeamMemberships(team))

	// Retrieve all users
	users, err := db.ListUsers(ctx, otf.UserListOptions{})
	require.NoError(t, err)

	require.Contains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)

	// Retrieve users in org
	users, err = db.ListUsers(ctx, otf.UserListOptions{Organization: otf.String(org.Name())})
	require.NoError(t, err)

	assert.NotContains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)

	// Retrieve users in org belonging to team
	users, err = db.ListUsers(ctx, otf.UserListOptions{
		Organization: otf.String(org.Name()),
		TeamName:     otf.String(team.Name()),
	})
	require.NoError(t, err)

	assert.NotContains(t, users, user1)
	assert.NotContains(t, users, user2)
	assert.Contains(t, users, user3)
}

func TestUser_Delete(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	user := CreateTestUser(t, db)

	err := db.DeleteUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	users, err := db.ListUsers(ctx, otf.UserListOptions{})
	require.NoError(t, err)
	assert.NotContains(t, users, user)
}
