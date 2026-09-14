package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/user"
	otfvariable "github.com/leg100/otf/internal/variable"
	otfworkspace "github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ReadWorkspacesPermission demonstrates that a team granted the
// organization-wide 'read workspaces' permission can view every workspace in
// the organization - along with its runs and state - but cannot modify a
// workspace nor start a run.
func TestIntegration_ReadWorkspacesPermission(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)
	ws := daemon.createWorkspace(t, ctx, org)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	run := daemon.createRun(t, ctx, ws, cv, nil)

	// Create user and add as member of viewers team, which is granted
	// organization-wide read access to workspaces.
	viewer := daemon.createUser(t)
	viewers := daemon.createTeam(t, ctx, org)
	err := daemon.Users.AddTeamMembership(ctx, viewers.ID, []user.Username{viewer.Username})
	require.NoError(t, err)
	_, err = daemon.Teams.UpdateTeam(ctx, viewers.ID, team.UpdateTeamOptions{
		OrganizationAccessOptions: team.OrganizationAccessOptions{
			ReadWorkspaces: internal.Ptr(true),
		},
	})
	require.NoError(t, err)
	// Refresh viewer user context to include new team membership
	_, viewerCtx := daemon.getUserCtx(t, adminCtx, viewer.Username)

	t.Run("list workspaces", func(t *testing.T) {
		got, err := daemon.Workspaces.ListWorkspaces(viewerCtx, otfworkspace.ListOptions{
			Organization: &org.Name,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(got.Items))
	})

	t.Run("get workspace", func(t *testing.T) {
		_, err := daemon.Workspaces.GetWorkspace(viewerCtx, ws.ID)
		require.NoError(t, err)
	})

	t.Run("list runs", func(t *testing.T) {
		got, err := daemon.Runs.ListRuns(viewerCtx, otfrun.ListOptions{
			WorkspaceID: &ws.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(got.Items))
	})

	t.Run("get run", func(t *testing.T) {
		_, err := daemon.Runs.GetRun(viewerCtx, run.ID)
		require.NoError(t, err)
	})

	t.Run("get current state version", func(t *testing.T) {
		// workspace has no state yet, but the viewer should get as far as a
		// not-found rather than being refused access.
		_, err := daemon.State.GetCurrentStateVersion(viewerCtx, ws.ID)
		assert.ErrorIs(t, err, internal.ErrResourceNotFound)
	})

	t.Run("cannot update workspace", func(t *testing.T) {
		_, err := daemon.Workspaces.UpdateWorkspace(viewerCtx, ws.ID, otfworkspace.UpdateOptions{
			Description: internal.Ptr("updated"),
		})
		assert.ErrorIs(t, err, internal.ErrAccessNotPermitted)
	})

	t.Run("cannot create run", func(t *testing.T) {
		_, err := daemon.Runs.CreateRun(viewerCtx, ws.ID, otfrun.CreateOptions{
			ConfigurationVersionID: &cv.ID,
		})
		assert.ErrorIs(t, err, internal.ErrAccessNotPermitted)
	})

	t.Run("cannot create variable", func(t *testing.T) {
		_, err := daemon.Variables.CreateVariable(viewerCtx, ws.ID, otfvariable.CreateVariableOptions{
			Key:      internal.Ptr("foo"),
			Value:    internal.Ptr("bar"),
			Category: internal.Ptr(otfvariable.CategoryTerraform),
		})
		assert.ErrorIs(t, err, internal.ErrAccessNotPermitted)
	})
}
