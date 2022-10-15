package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspacePermissions_Set(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	team := createTestTeam(t, db, org)

	err := db.SetWorkspacePermission(ctx, ws.SpecName(), team.Name(), otf.WorkspacePlanRole)
	require.NoError(t, err)

	t.Run("Update", func(t *testing.T) {
		err := db.SetWorkspacePermission(ctx, ws.SpecName(), team.Name(), otf.WorkspaceAdminRole)
		require.NoError(t, err)
	})
}

func TestWorkspacePermissions_List(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	team1 := createTestTeam(t, db, org)
	team2 := createTestTeam(t, db, org)
	_ = createTestWorkspacePermission(t, db, ws, team1, otf.WorkspaceAdminRole)
	_ = createTestWorkspacePermission(t, db, ws, team2, otf.WorkspacePlanRole)

	perms, err := db.ListWorkspacePermissions(ctx, ws.SpecName())
	require.NoError(t, err)
	assert.Equal(t, 2, len(perms))
	// TODO: compare contents of listing - we cannot do this yet because the team
	// obj includes the organization name whereas the team obj within a
	// workspace permission does not, thus comparing the contents will fail.
	// Once the planned work to remove organization name from team is complete
	// we should complete this test.
}

func TestWorkspacePermissions_Unset(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	team := createTestTeam(t, db, org)
	perm := createTestWorkspacePermission(t, db, ws, team, otf.WorkspaceAdminRole)

	perms, err := db.ListWorkspacePermissions(ctx, ws.SpecName())
	require.NoError(t, err)
	assert.Equal(t, 1, len(perms))
	assert.Equal(t, perm.Permission, perms[0].Permission)

	err = db.UnsetWorkspacePermission(ctx, ws.SpecName(), team.Name())
	require.NoError(t, err)

	perms, err = db.ListWorkspacePermissions(ctx, ws.SpecName())
	require.NoError(t, err)
	assert.Equal(t, 0, len(perms))
}
