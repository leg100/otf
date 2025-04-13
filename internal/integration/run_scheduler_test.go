package integration

import (
	"testing"
	"time"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunScheduler(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, &config{})
	user := userFromContext(t, ctx)

	// watch workspace events
	workspaceEvents, unsub := daemon.Workspaces.Watch(ctx)
	defer unsub()

	ws := daemon.createWorkspace(t, ctx, nil)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	run1 := daemon.createRun(t, ctx, ws, cv, nil)
	run2 := daemon.createRun(t, ctx, ws, cv, nil)

	// Wait for Run#1 to lock workspace
	waitWorkspaceLock(t, workspaceEvents, run1.ID)

	// Wait for Run#1 to be planned
	daemon.waitRunStatus(t, run1.ID, runstatus.Planned)
	// Run#2 should still be pending
	assert.Equal(t, runstatus.Pending, daemon.getRun(t, ctx, run2.ID).Status)

	// Apply Run#1
	err := daemon.Runs.Apply(ctx, run1.ID)
	require.NoError(t, err)

	// Wait for Run#1 to be applied
	daemon.waitRunStatus(t, run1.ID, runstatus.Applied)

	// Wait for Run#2 to lock workspace
	waitWorkspaceLock(t, workspaceEvents, run2.ID)

	// Wait for Run#2 to be planned&finished (because there are no changes)
	daemon.waitRunStatus(t, run2.ID, runstatus.PlannedAndFinished)

	// Wait for workspace to be unlocked
	waitWorkspaceLock(t, workspaceEvents, nil)

	// User locks workspace
	_, err = daemon.Workspaces.Lock(ctx, ws.ID, nil)
	require.NoError(t, err)

	// Create another run, it should remain in pending status.
	run3 := daemon.createRun(t, ctx, ws, cv, nil)

	// Workspace should still be locked by user
	waitWorkspaceLock(t, workspaceEvents, user.Username)

	// User unlocks workspace
	_, err = daemon.Workspaces.Unlock(ctx, ws.ID, nil, false)
	require.NoError(t, err)

	// Run #3 should now proceed to planned&finished
	daemon.waitRunStatus(t, run3.ID, runstatus.PlannedAndFinished)
}

func waitWorkspaceLock(t *testing.T, events <-chan pubsub.Event[*workspace.Workspace], lock resource.ID) {
	t.Helper()

	timeout := time.After(5 * time.Second)
	for {
		select {
		case event := <-events:
			if lock != nil {
				if event.Payload.Lock != nil && lock == event.Payload.Lock {
					return
				}
			} else {
				if event.Payload.Lock == nil {
					return
				}
			}
		case <-timeout:
			t.Fatalf("timed out waiting for workspace lock condition")
		}
	}
}
