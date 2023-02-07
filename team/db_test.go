package team

import (
	"context"
	"testing"

	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	teamDB := newDB(db)
	org := organization.CreateTestOrganization(t, db)
	team := newTeam("team-awesome", org.Name())

	defer teamDB.DeleteTeam(ctx, team.ID())

	err := teamDB.CreateTeam(ctx, team)
	require.NoError(t, err)
}

func TestTeam_Update_ByID(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	teamDB := newDB(db)

	org := organization.CreateTestOrganization(t, db)
	team := CreateTestTeam(t, teamDB, org.Name())

	_, err := teamDB.UpdateTeam(ctx, team.ID(), func(team *Team) error {
		return team.Update(UpdateTeamOptions{
			OrganizationAccess: OrganizationAccess{
				ManageWorkspaces: true,
				ManageVCS:        true,
				ManageRegistry:   true,
			},
		})
	})
	require.NoError(t, err)

	got, err := teamDB.GetTeam(ctx, team.Name(), org.Name())
	require.NoError(t, err)

	assert.True(t, got.OrganizationAccess().ManageWorkspaces)
	assert.True(t, got.OrganizationAccess().ManageVCS)
	assert.True(t, got.OrganizationAccess().ManageRegistry)
}

func TestTeam_Get(t *testing.T) {
	db := sql.NewTestDB(t)
	teamDB := newDB(db)

	org := organization.CreateTestOrganization(t, db)
	team := CreateTestTeam(t, teamDB, org.Name())

	got, err := teamDB.GetTeam(context.Background(), team.Name(), org.Name())
	require.NoError(t, err)

	assert.Equal(t, team, got)
}

func TestTeam_GetByID(t *testing.T) {
	db := sql.NewTestDB(t)
	teamDB := newDB(db)

	org := organization.CreateTestOrganization(t, db)
	want := CreateTestTeam(t, db, org.Name())

	got, err := teamDB.GetTeamByID(context.Background(), want.ID())
	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func TestTeam_List(t *testing.T) {
	db := sql.NewTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	teamDB := newDB(db)

	team1 := CreateTestTeam(t, db, org.Name())
	team2 := CreateTestTeam(t, db, org.Name())
	team3 := CreateTestTeam(t, db, org.Name())

	got, err := teamDB.ListTeams(context.Background(), org.Name())
	require.NoError(t, err)

	assert.Contains(t, got, team1)
	assert.Contains(t, got, team2)
	assert.Contains(t, got, team3)
}
