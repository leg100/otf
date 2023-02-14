package auth

import (
	"context"
	"testing"

	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam_ListTeamMembers(t *testing.T) {
	db := newTestDB(t)

	org := organization.CreateTestOrganization(t, db)
	team := CreateTestTeam(t, db, org.Name())

	memberships := []newUserOption{
		WithOrganizations(org.Name()),
		WithTeams(team),
	}
	user1 := createTestUser(t, db, memberships...)
	user2 := createTestUser(t, db, memberships...)

	got, err := db.listTeamMembers(context.Background(), team.ID())
	require.NoError(t, err)

	assert.Contains(t, got, user1)
	assert.Contains(t, got, user2)
}

func TestTeam_Create(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	org := organization.CreateTestOrganization(t, db)
	team := newTeam(createTeamOptions{"team-awesome", org.Name()})

	defer db.deleteTeam(ctx, team.ID())

	err := db.createTeam(ctx, team)
	require.NoError(t, err)
}

func TestTeam_Update_ByID(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	org := organization.CreateTestOrganization(t, db)
	team := CreateTestTeam(t, db, org.Name())

	_, err := db.UpdateTeam(ctx, team.ID(), func(team *Team) error {
		return team.Update(UpdateTeamOptions{
			OrganizationAccess: OrganizationAccess{
				ManageWorkspaces: true,
				ManageVCS:        true,
				ManageRegistry:   true,
			},
		})
	})
	require.NoError(t, err)

	got, err := db.getTeam(ctx, team.Name(), org.Name())
	require.NoError(t, err)

	assert.True(t, got.OrganizationAccess().ManageWorkspaces)
	assert.True(t, got.OrganizationAccess().ManageVCS)
	assert.True(t, got.OrganizationAccess().ManageRegistry)
}

func TestTeam_Get(t *testing.T) {
	db := newTestDB(t)

	org := organization.CreateTestOrganization(t, db)
	team := CreateTestTeam(t, db, org.Name())

	got, err := db.getTeam(context.Background(), team.Name(), org.Name())
	require.NoError(t, err)

	assert.Equal(t, team, got)
}

func TestTeam_GetByID(t *testing.T) {
	db := newTestDB(t)

	org := organization.CreateTestOrganization(t, db)
	want := CreateTestTeam(t, db, org.Name())

	got, err := db.getTeamByID(context.Background(), want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestTeam_List(t *testing.T) {
	db := newTestDB(t)
	org := organization.CreateTestOrganization(t, db)

	team1 := CreateTestTeam(t, db, org.Name())
	team2 := CreateTestTeam(t, db, org.Name())
	team3 := CreateTestTeam(t, db, org.Name())

	got, err := db.listTeams(context.Background(), org.Name())
	require.NoError(t, err)

	assert.Contains(t, got, team1)
	assert.Contains(t, got, team2)
	assert.Contains(t, got, team3)
}
