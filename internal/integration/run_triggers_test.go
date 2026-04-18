package integration

import (
	"testing"

	"github.com/leg100/otf/internal/runstatus"
	"github.com/stretchr/testify/require"
)

// TestRunTriggers tests run trigger functionality
func TestRunTriggers(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t)

	// Create two workspaces
	ws1 := daemon.createWorkspace(t, ctx, nil)
	ws2 := daemon.createWorkspace(t, ctx, nil)

	// Create trigger with ws1 as the triggering workspace and ws2 as the
	// triggered workspace.
	_, err := daemon.RunTriggers.CreateRunTrigger(ctx, ws2.ID, ws1.ID)
	require.NoError(t, err)

	// The triggered workspace first needs some config.
	_ = daemon.createAndUploadConfigurationVersion(t, ctx, ws2, nil)

	// Create a run on the triggering workspace
	cv1 := daemon.createAndUploadConfigurationVersion(t, ctx, ws1, nil)
	_ = daemon.createRun(t, ctx, ws1, cv1, nil)

	// Wait for run to be applied on the triggering workspace.
	for re := range daemon.runEvents {
		require.True(t, re.Payload.WorkspaceID == ws1.ID)

		switch re.Payload.Status {
		case runstatus.Planned:
			err := daemon.Runs.ApplyRun(ctx, re.Payload.ID)
			require.NoError(t, err)
		case runstatus.Applied:
			goto done
		}
	}
done:

	// Run should now be automatically created on the triggered workspace.
	for re := range daemon.runEvents {
		require.True(t, re.Payload.WorkspaceID == ws2.ID)

		switch re.Payload.Status {
		case runstatus.Planned:
			err := daemon.Runs.ApplyRun(ctx, re.Payload.ID)
			require.NoError(t, err)
		case runstatus.Applied:
			return
		}
	}
}
