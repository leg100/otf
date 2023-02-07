package team

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam_Create(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	team := otf.NewTeam("team-awesome", org)

	defer db.DeleteTeam(ctx, team.ID())

	err := db.CreateTeam(ctx, team)
	require.NoError(t, err)
}

func TestTeam_Update_ByID(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)

	org := CreateTestOrganization(t, db)
	team := createTestTeam(t, db, org)

	_, err := db.UpdateTeam(ctx, team.ID(), func(team *otf.Team) error {
		return team.Update(otf.UpdateTeamOptions{
			OrganizationAccess: otf.OrganizationAccess{
				ManageWorkspaces: true,
				ManageVCS:        true,
				ManageRegistry:   true,
			},
		})
	})
	require.NoError(t, err)

	got, err := db.GetTeam(ctx, team.Name(), org.Name())
	require.NoError(t, err)

	assert.True(t, got.OrganizationAccess().ManageWorkspaces)
	assert.True(t, got.OrganizationAccess().ManageVCS)
	assert.True(t, got.OrganizationAccess().ManageRegistry)
}

func TestTeam_Get(t *testing.T) {
	db := NewTestDB(t)

	org := CreateTestOrganization(t, db)
	team := createTestTeam(t, db, org)

	got, err := db.GetTeam(context.Background(), team.Name(), org.Name())
	require.NoError(t, err)

	assert.Equal(t, team, got)
}

func TestTeam_GetByID(t *testing.T) {
	db := NewTestDB(t)

	org := CreateTestOrganization(t, db)
	want := createTestTeam(t, db, org)

	got, err := db.GetTeamByID(context.Background(), want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestTeam_List(t *testing.T) {
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	team1 := createTestTeam(t, db, org)
	team2 := createTestTeam(t, db, org)
	team3 := createTestTeam(t, db, org)

	got, err := db.ListTeams(context.Background(), org.Name())
	require.NoError(t, err)

	assert.Contains(t, got, team1)
	assert.Contains(t, got, team2)
	assert.Contains(t, got, team3)
}

func TestTeam_ListTeamMembers(t *testing.T) {
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	team := createTestTeam(t, db, org)
	memberships := []otf.NewUserOption{
		otf.WithOrganizationMemberships(org.Name()),
		otf.WithTeamMemberships(team),
	}
	user1 := CreateTestUser(t, db, memberships...)
	user2 := CreateTestUser(t, db, memberships...)

	got, err := db.ListTeamMembers(context.Background(), team.ID())
	require.NoError(t, err)

	assert.Contains(t, got, user1)
	assert.Contains(t, got, user2)
}
