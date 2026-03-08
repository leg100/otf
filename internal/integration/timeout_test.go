package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Timeout demonstrates timing out a run that remains in the
// planning state beyond the specified planning timeout.
func TestIntegration_Timeout(t *testing.T) {
	integrationTest(t)

	// Set a very low planning timeout and check very frequently.
	daemon, org, ctx := setup(t, withTimeouts(
		time.Second,
		time.Second,
		100*time.Millisecond,
	))
	ws := daemon.createWorkspace(t, ctx, org)

	// Setup a http server, to which the terraform 'http' data source will
	// connect, causing it to hang, thereby keeping OTF run in the planning
	// state.
	done := make(chan struct{})
	t.Cleanup(func() {
		done <- struct{}{}
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only return once test completes
		<-done
	}))

	// create some config and upload
	cv := daemon.createConfigurationVersion(t, ctx, ws, nil)
	config := newRootModule(t, daemon.System.Hostname(), org.Name, ws.Name, fmt.Sprintf(`
data "http" "wait" {
	url = "%s"
}
`, srv.URL))
	tarball, err := internal.Pack(config)
	require.NoError(t, err)
	err = daemon.Configs.UploadConfig(ctx, cv.ID, tarball)
	require.NoError(t, err)

	// create run and wait for it to enter canceled state
	run := daemon.createRun(t, ctx, ws, cv, nil)
	run = daemon.waitRunStatus(t, ctx, run.ID, runstatus.Canceled)

	// run should have reached planning state before being timed out and being
	// forced into a canceled state.
	_, err = run.StatusTimestamp(runstatus.Planning)
	assert.NoError(t, err)
	assert.Equal(t, runstatus.Canceled, run.Status)
}
