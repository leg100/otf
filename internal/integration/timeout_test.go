package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/daemon"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Timeout demonstrates timing out a run that remains in the
// planning state beyond the specified planning timeout.
func TestIntegration_Timeout(t *testing.T) {
	integrationTest(t)

	// Set a very low planning timeout and check very frequently.
	svc, org, ctx := setup(t, &config{
		Config: daemon.Config{
			PlanningTimeout:              time.Second,
			OverrideTimeoutCheckInterval: 100 * time.Millisecond,
		},
	})
	ws := svc.createWorkspace(t, ctx, org)

	// watch run events
	runsSub, runsUnsub := svc.Runs.Watch(ctx)
	defer runsUnsub()

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
	cv := svc.createConfigurationVersion(t, ctx, ws, nil)
	config := newRootModule(t, svc.System.Hostname(), org.Name, ws.Name, fmt.Sprintf(`
data "http" "wait" {
	url = "%s"
}
`, srv.URL))
	tarball, err := internal.Pack(config)
	require.NoError(t, err)
	err = svc.Configs.UploadConfig(ctx, cv.ID, tarball)
	require.NoError(t, err)

	// create run and wait for it to finish
	run := svc.createRun(t, ctx, ws, cv)
	for event := range runsSub {
		if event.Payload.ID != run.ID {
			continue
		}
		if event.Payload.Done() {
			run = event.Payload
			break
		}
	}
	// run should have reached planning state before being timed out and being
	// forced into a canceled state.
	_, err = run.StatusTimestamp(otfrun.RunPlanning)
	assert.NoError(t, err)
	assert.Equal(t, otfrun.RunCanceled, run.Status)
}
