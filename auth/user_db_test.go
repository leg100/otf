package auth_test

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Get(t *testing.T) {
	org1 := organization.CreateTestOrganization(t, db)
	org2 := organization.CreateTestOrganization(t, db)
	team1 := CreateTestTeam(t, db, org1.Name())
	team2 := CreateTestTeam(t, db, org2.Name())

	user := createTestUser(t, db,
		withOrganizations(org1.Name(), org2.Name()),
		withTeams(team1, team2))

	session1 := createTestSession(t, db, user.ID(), nil)
	_ = createTestSession(t, db, user.ID(), nil)

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
			got, err := db.getUser(context.Background(), tt.spec)
			require.NoError(t, err)

			assert.Equal(t, got.ID(), user.ID())
			assert.Equal(t, got.Username(), user.Username())
			assert.Equal(t, got.CreatedAt(), user.CreatedAt())
			assert.Equal(t, got.UpdatedAt(), user.UpdatedAt())
			assert.Equal(t, 2, len(got.Organizations()))
			assert.Equal(t, 2, len(got.teams))
		})
	}
}

func TestUser_Get_NotFound(t *testing.T) {
	db := newTestDB(t)

	_, err := db.getUser(context.Background(), otf.UserSpec{Username: otf.String("does-not-exist")})
	assert.Equal(t, otf.ErrResourceNotFound, err)
}

func TestUser_List(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	org := organization.CreateTestOrganization(t, db)
	user1 := createTestUser(t, db)
	user2 := createTestUser(t, db, withOrganizations(org.Name()))
	user3 := createTestUser(t, db, withOrganizations(org.Name()))

	// Retrieve all users
	users, err := db.listUsers(ctx, org.Name())
	require.NoError(t, err)

	require.Contains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)

	// Retrieve users in org
	users, err = db.listUsers(ctx, org.Name())
	require.NoError(t, err)

	assert.NotContains(t, users, user1)
	assert.Contains(t, users, user2)
	assert.Contains(t, users, user3)
}

func TestUser_Delete(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	user := createTestUser(t, db)

	spec := otf.UserSpec{Username: otf.String(user.Username())}
	err := db.DeleteUser(ctx, spec)
	require.NoError(t, err)

	_, err = db.getUser(ctx, spec)
	assert.Equal(t, err, otf.ErrResourceNotFound)
}

func TestUser_AddTeamMembership(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	org := organization.CreateTestOrganization(t, db)
	team := CreateTestTeam(t, db, org.Name())
	user := createTestUser(t, db, withOrganizations(org.Name()))

	err := db.addTeamMembership(ctx, user.ID(), team.ID())
	require.NoError(t, err)

	got, err := db.getUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.Contains(t, got.teams, team)
}

func TestUser_RemoveTeamMembership(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	org := organization.CreateTestOrganization(t, db)
	team := CreateTestTeam(t, db, org.Name())
	user := createTestUser(t, db, withOrganizations(org.Name()), withTeams(team))

	err := db.removeTeamMembership(ctx, user.ID(), team.ID())
	require.NoError(t, err)

	got, err := db.getUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
	require.NoError(t, err)

	assert.NotContains(t, got.teams, team)
}
