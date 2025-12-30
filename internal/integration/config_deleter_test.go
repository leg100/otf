package integration

import (
	"testing"
	"time"

	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

// TestConfigDeleter tests the config deleter subsystem, which deletes configs
// older than a user-specified time period.
func TestConfigDeleter(t *testing.T) {
	integrationTest(t)

	// Delete configs older than 1 hour, and check configs every second.
	daemon, org, ctx := setup(t, withDeleteConfigsAfter(time.Hour, time.Second))

	yesterday := time.Now().Add(-time.Hour * 24)

	ws := daemon.createWorkspace(t, ctx, org)

	// Create two old configs that should be deleted, and two new configs that
	// should not be deleted.
	cv1 := daemon.createConfigurationVersion(t, ctx, ws, &configversion.CreateOptions{CreatedAt: &yesterday})
	cv2 := daemon.createConfigurationVersion(t, ctx, ws, &configversion.CreateOptions{CreatedAt: &yesterday})
	cv3 := daemon.createConfigurationVersion(t, ctx, ws, nil)
	cv4 := daemon.createConfigurationVersion(t, ctx, ws, nil)

	// Also attach such to each config to be deleted, to make it more realistic
	// and check that cascades are working.
	_ = daemon.createRun(t, ctx, ws, cv1, nil)
	_ = daemon.createRun(t, ctx, ws, cv2, nil)

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("configs were not deleted within timeout")
		case <-ticker.C:
			configs, err := daemon.Configs.List(ctx, ws.ID, configversion.ListOptions{})
			require.NoError(t, err)
			if len(configs.Items) == 2 {
				// Good, only 2 configs remaining. Check that they are the new
				// configs by checking that they both still exist.
				_, err := daemon.Configs.Get(ctx, cv3.ID)
				require.NoError(t, err)
				_, err = daemon.Configs.Get(ctx, cv4.ID)
				require.NoError(t, err)

				// Check that both runs belonging to the deleted configs have
				// been deleted as well.
				// Listing runs site-wide requires site admin user
				runs, err := daemon.Runs.List(adminCtx, run.ListOptions{})
				require.NoError(t, err)
				require.Equal(t, 0, len(runs.Items))

				// Tests pass.
				return
			}
		}
	}
}
