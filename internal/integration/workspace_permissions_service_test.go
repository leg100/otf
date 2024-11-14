package integration

import (
	"errors"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_WorkspacePermissionsService(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	t.Run("set permission", func(t *testing.T) {
		ws := svc.createWorkspace(t, ctx, org)
		team := svc.createTeam(t, ctx, org)
		err := svc.Workspaces.SetPermission(ctx, ws.ID, team.ID, rbac.WorkspacePlanRole)
		require.NoError(t, err)
	})

	t.Run("unset permission", func(t *testing.T) {
		ws := svc.createWorkspace(t, ctx, org)
		team := svc.createTeam(t, ctx, org)
		err := svc.Workspaces.SetPermission(ctx, ws.ID, team.ID, rbac.WorkspacePlanRole)
		require.NoError(t, err)

		err = svc.Workspaces.UnsetPermission(ctx, ws.ID, team.ID)
		require.NoError(t, err)

		policy, err := svc.Workspaces.GetWorkspacePolicy(ctx, ws.ID)
		require.NoError(t, err)
		assert.Empty(t, policy.Permissions)
	})

	t.Run("get policy", func(t *testing.T) {
		ws := svc.createWorkspace(t, ctx, org)
		scum := svc.createTeam(t, ctx, org)
		skates := svc.createTeam(t, ctx, org)
		cherries := svc.createTeam(t, ctx, org)
		err := svc.Workspaces.SetPermission(ctx, ws.ID, scum.ID, rbac.WorkspaceAdminRole)
		require.NoError(t, err)
		err = svc.Workspaces.SetPermission(ctx, ws.ID, skates.ID, rbac.WorkspaceReadRole)
		require.NoError(t, err)
		err = svc.Workspaces.SetPermission(ctx, ws.ID, cherries.ID, rbac.WorkspacePlanRole)
		require.NoError(t, err)

		got, err := svc.Workspaces.GetWorkspacePolicy(ctx, ws.ID)
		require.NoError(t, err)

		assert.Equal(t, 3, len(got.Permissions))
		assert.Contains(t, got.Permissions, authz.WorkspacePermission{
			TeamID: scum.ID,
			Role:   rbac.WorkspaceAdminRole,
		})
		assert.Contains(t, got.Permissions, authz.WorkspacePermission{
			TeamID: skates.ID,
			Role:   rbac.WorkspaceReadRole,
		})
		assert.Contains(t, got.Permissions, authz.WorkspacePermission{
			TeamID: cherries.ID,
			Role:   rbac.WorkspacePlanRole,
		})
	})

	t.Run("workspace not found", func(t *testing.T) {
		nonExistentID := resource.NewID(resource.WorkspaceKind)
		_, err := svc.Workspaces.GetWorkspacePolicy(ctx, nonExistentID)
		require.True(t, errors.Is(err, internal.ErrResourceNotFound))
	})
}
