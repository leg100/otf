package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/releases"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunCancel demonstrates a run being canceled mid-flow.
func TestIntegration_RunCancel(t *testing.T) {
	integrationTest(t)

	// stage a fake terraform bin that sleeps until it receives an interrupt
	// signal
	bins := filepath.Join(t.TempDir(), "bins")
	dst := filepath.Join(bins, releases.DefaultTerraformVersion, "terraform")
	err := os.MkdirAll(filepath.Dir(dst), 0o755)
	require.NoError(t, err)
	wd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Symlink(filepath.Join(wd, "testdata/cancelme"), dst)
	require.NoError(t, err)

	daemon, org, ctx := setup(t, &config{terraformBinDir: dst})

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
	daemon.startAgent(t, ctx, org.Name, "", agent.Config{TerraformBinDir: bins})

	// create workspace specifying that it use an external agent.
	ws, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:          internal.String("ws-1"),
		Organization:  internal.String(org.Name),
		ExecutionMode: workspace.ExecutionModePtr(workspace.AgentExecutionMode),
	})
	require.NoError(t, err)

	// create a variable so that the fake bin knows the url of the temp http
	// server
	_, err = daemon.CreateWorkspaceVariable(ctx, ws.ID, variable.CreateVariableOptions{
		Key:      internal.String("URL"),
		Value:    internal.String(srv.URL),
		Category: variable.VariableCategoryPtr(variable.CategoryEnv),
	})
	require.NoError(t, err)

	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	r := daemon.createRun(t, ctx, ws, cv)

	// fake bin process has started
	require.Equal(t, "started", <-got)

	// we can now send interrupt
	err = daemon.Cancel(ctx, r.ID)
	require.NoError(t, err)

	// fake bin has received interrupt
	require.Equal(t, "canceled", <-got)
}
