package team

import (
	"context"
	"testing"

	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
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

func TestWorkspacePermissions_Set(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	team := createTestTeam(t, db, org)

	err := db.SetWorkspacePermission(ctx, ws.ID(), team.Name(), rbac.WorkspacePlanRole)
	require.NoError(t, err)

	t.Run("Update", func(t *testing.T) {
		err := db.SetWorkspacePermission(ctx, ws.ID(), team.Name(), rbac.WorkspaceAdminRole)
		require.NoError(t, err)
	})
}

func TestWorkspacePermissions_List(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	team1 := createTestTeam(t, db, org)
	team2 := createTestTeam(t, db, org)
	perm1 := createTestWorkspacePermission(t, db, ws, team1, rbac.WorkspaceAdminRole)
	perm2 := createTestWorkspacePermission(t, db, ws, team2, rbac.WorkspacePlanRole)

	perms, err := db.ListWorkspacePermissions(ctx, ws.ID())
	require.NoError(t, err)
	if assert.Equal(t, 2, len(perms)) {
		assert.Contains(t, perms, perm1)
		assert.Contains(t, perms, perm2)
	}
}

func TestWorkspacePermissions_Unset(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	team := createTestTeam(t, db, org)
	_ = createTestWorkspacePermission(t, db, ws, team, rbac.WorkspaceAdminRole)

	err := db.UnsetWorkspacePermission(ctx, ws.ID(), team.Name())
	require.NoError(t, err)

	perms, err := db.ListWorkspacePermissions(ctx, ws.ID())
	require.NoError(t, err)
	assert.Equal(t, 0, len(perms))
}
