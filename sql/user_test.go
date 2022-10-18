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
	ctx := context.Background()

	org := createTestOrganization(t, db)
	user := createTestUser(t, db)

	err := db.AddOrganizationMembership(ctx, user.ID(), org.ID())
	require.NoError(t, err)

	got, err := db.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Contains(t, got.Organizations(), org)
}

func TestUser_RemoveOrganizationMembership(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	org := createTestOrganization(t, db)
	user := createTestUser(t, db, otf.WithOrganizationMemberships(org))

	err := db.RemoveOrganizationMembership(ctx, user.ID(), org.ID())
	require.NoError(t, err)

	got, err := db.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.NotContains(t, got.Organizations(), org)
}

func TestUser_AddTeamMembership(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	org := createTestOrganization(t, db)
	team := createTestTeam(t, db, org)
	user := createTestUser(t, db, otf.WithOrganizationMemberships(org))

	err := db.AddTeamMembership(ctx, user.ID(), team.ID())
	require.NoError(t, err)

	got, err := db.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Contains(t, got.Teams(), team)
}

func TestUser_RemoveTeamMembership(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	org := createTestOrganization(t, db)
	team := createTestTeam(t, db, org)
	user := createTestUser(t, db, otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))

	err := db.RemoveTeamMembership(ctx, user.ID(), team.ID())
	require.NoError(t, err)

	got, err := db.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.NotContains(t, got.Teams(), team)
}

func TestUser_Get(t *testing.T) {
	db := newTestDB(t)

	org1 := createTestOrganization(t, db)
	org2 := createTestOrganization(t, db)
	team1 := createTestTeam(t, db, org1)
	team2 := createTestTeam(t, db, org2)

	user := createTestUser(t, db,
		otf.WithOrganizationMemberships(org1, org2),
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
			spec: otf.UserSpec{SessionToken: &session1.Token},
		},
		{
			name: "auth token ID",
			spec: otf.UserSpec{AuthenticationTokenID: otf.String(token1.ID())},
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
			assert.Equal(t, 2, len(got.Sessions()))
			assert.Equal(t, 2, len(got.Tokens()))
			assert.Equal(t, 2, len(got.Teams()))
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

	assert.Equal(t, 2, len(got.Sessions()))
}

func TestUser_List(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	team := createTestTeam(t, db, org)
	user1 := createTestUser(t, db)
	user2 := createTestUser(t, db, otf.WithOrganizationMemberships(org))
	user3 := createTestUser(t, db, otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))

	// Retrieve all users
	users, err := db.ListUsers(ctx, otf.UserListOptions{})
	require.NoError(t, err)

	assert.Contains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)

	// Retrieve users in org
	users, err = db.ListUsers(ctx, otf.UserListOptions{OrganizationName: otf.String(org.Name())})
	require.NoError(t, err)

	assert.NotContains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)

	// Retrieve users in org belonging to team
	users, err = db.ListUsers(ctx, otf.UserListOptions{
		OrganizationName: otf.String(org.Name()),
		TeamName:         otf.String(team.Name()),
	})
	require.NoError(t, err)

	assert.NotContains(t, users, user1)
	assert.NotContains(t, users, user2)
	assert.Contains(t, users, user3)
}

func TestUser_Delete(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	user := createTestUser(t, db)

	err := db.DeleteUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	users, err := db.ListUsers(ctx, otf.UserListOptions{})
	require.NoError(t, err)
	assert.NotContains(t, users, user)
}
