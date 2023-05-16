package integration

import (
	"errors"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_WorkspacePermissionsService(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)
	org := svc.createOrganization(t, ctx)

	t.Run("set permission", func(t *testing.T) {
		ws := svc.createWorkspace(t, ctx, org)
		team := svc.createTeam(t, ctx, org)
		err := svc.SetPermission(ctx, ws.ID, team.Name, rbac.WorkspacePlanRole)
		require.NoError(t, err)
	})

	t.Run("unset permission", func(t *testing.T) {
		ws := svc.createWorkspace(t, ctx, org)
		team := svc.createTeam(t, ctx, org)
		err := svc.SetPermission(ctx, ws.ID, team.Name, rbac.WorkspacePlanRole)
		require.NoError(t, err)

		err = svc.UnsetPermission(ctx, ws.ID, team.Name)
		require.NoError(t, err)

		policy, err := svc.GetPolicy(ctx, ws.ID)
		require.NoError(t, err)
		assert.Empty(t, policy.Permissions)
	})

	t.Run("get policy", func(t *testing.T) {
		ws := svc.createWorkspace(t, ctx, org)
		scum := svc.createTeam(t, ctx, org)
		skates := svc.createTeam(t, ctx, org)
		cherries := svc.createTeam(t, ctx, org)
		err := svc.SetPermission(ctx, ws.ID, scum.Name, rbac.WorkspaceAdminRole)
		require.NoError(t, err)
		err = svc.SetPermission(ctx, ws.ID, skates.Name, rbac.WorkspaceReadRole)
		require.NoError(t, err)
		err = svc.SetPermission(ctx, ws.ID, cherries.Name, rbac.WorkspacePlanRole)
		require.NoError(t, err)

		got, err := svc.GetPolicy(ctx, ws.ID)
		require.NoError(t, err)

		assert.Equal(t, org.Name, got.Organization)
		assert.Equal(t, ws.ID, got.WorkspaceID)
		assert.Equal(t, 3, len(got.Permissions))
		assert.Contains(t, got.Permissions, internal.WorkspacePermission{
			Team:   scum.Name,
			TeamID: scum.ID,
			Role:   rbac.WorkspaceAdminRole,
		})
		assert.Contains(t, got.Permissions, internal.WorkspacePermission{
			Team:   skates.Name,
			TeamID: skates.ID,
			Role:   rbac.WorkspaceReadRole,
		})
		assert.Contains(t, got.Permissions, internal.WorkspacePermission{
			Team:   cherries.Name,
			TeamID: cherries.ID,
			Role:   rbac.WorkspacePlanRole,
		})
	})

	t.Run("workspace not found", func(t *testing.T) {
		_, err := svc.GetPolicy(ctx, "non-existent")
		require.True(t, errors.Is(err, internal.ErrResourceNotFound))
	})
}
