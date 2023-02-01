package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
