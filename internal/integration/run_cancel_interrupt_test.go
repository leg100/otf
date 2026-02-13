package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunCancelInterrupt tests cancelling a run via an interrupt
// signal, which occurs when a run is in the planning or applying state.
func TestIntegration_RunCancelInterrupt(t *testing.T) {
	integrationTest(t)

	// stage a fake terraform bin that sleeps until it receives an interrupt
	// signal
	bins := filepath.Join(t.TempDir(), "bins")
	dst := filepath.Join(bins, engine.Default.DefaultVersion(), "terraform")
	err := os.MkdirAll(filepath.Dir(dst), 0o755)
	require.NoError(t, err)
	wd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Symlink(filepath.Join(wd, "testdata/cancelme"), dst)
	require.NoError(t, err)

	daemon, org, ctx := setup(t)

	// run a temporary http server as a means of communicating with the fake
	// bin
	got := make(chan string)
	mux := http.NewServeMux()
	mux.HandleFunc("/started", func(w http.ResponseWriter, r *http.Request) {
		// fake bin has started
		got <- "started"
	})
	mux.HandleFunc("/canceled", func(w http.ResponseWriter, r *http.Request) {
		// fake bin has received interrupt signal
		got <- "canceled"
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	// start an external agent (it's the only way to specify a separate bin
	// directory currently).
	agent, _ := daemon.startAgent(t, ctx, org.Name, nil, "", withEngineBinDir(bins))

	// create workspace specifying that it use an external agent.
	ws, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
		Name:          new("ws-1"),
		Organization:  &org.Name,
		ExecutionMode: internal.Ptr(workspace.AgentExecutionMode),
		AgentPoolID:   &agent.AgentPool.ID,
	})
	require.NoError(t, err)

	// create a variable so that the fake bin knows the url of the temp http
	// server
	_, err = daemon.Variables.CreateWorkspaceVariable(ctx, ws.ID, variable.CreateVariableOptions{
		Key:      new("URL"),
		Value:    new(srv.URL),
		Category: internal.Ptr(variable.CategoryEnv),
	})
	require.NoError(t, err)

	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	r := daemon.createRun(t, ctx, ws, cv, nil)

	// fake bin process has started
	require.Equal(t, "started", <-got)

	// we can now send interrupt
	err = daemon.Runs.Cancel(ctx, r.ID)
	require.NoError(t, err)

	// fake bin has received interrupt
	require.Equal(t, "canceled", <-got)

	// canceling the job should result in the run then entering the canceled
	// state.
	daemon.waitRunStatus(t, ctx, r.ID, runstatus.Canceled)
}
