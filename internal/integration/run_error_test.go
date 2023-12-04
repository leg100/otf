package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

// TestRunError demonstrates a run failing with an error and checks that the
// error is correctly reported. The tests are run both for a run executed via
// the daemon, and for a run executed via an agent. They each use different
// mechanisms for reporting the error and we want to test both (the agent uses
// RPC calls whereas the daemon is in-process).
func TestRunError(t *testing.T) {
	integrationTest(t)

	// create a daemon and start an agent
	daemon, org, ctx := setup(t, nil)
	poolID := daemon.startAgent(t, ctx, org.Name, "", agent.Config{})

	// two tests: one run on the daemon, one via the agent.
	tests := []struct {
		name   string
		mode   workspace.ExecutionMode
		poolID *string
	}{
		{
			"execute run via daemon", workspace.RemoteExecutionMode, nil,
		},
		{
			"execute run via agent", workspace.AgentExecutionMode, &poolID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create workspace
			ws, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
				Name:          internal.String("ws-" + string(tt.mode)),
				Organization:  internal.String(org.Name),
				ExecutionMode: workspace.ExecutionModePtr(tt.mode),
				AgentPoolID:   tt.poolID,
			})
			require.NoError(t, err)

			// create some invalid config
			config := fmt.Sprintf(`
		terraform {
		  cloud {
			hostname = "%s"
			organization = "%s"

			workspaces {
			  name = "%s"
			}
		  }
		}
		# should be 'null_resource'
		resource "null_resourc" "e2e" {}
		`, daemon.Hostname(), org.Name, ws.Name)

			// upload config
			cv := daemon.createConfigurationVersion(t, ctx, ws, nil)
			path := t.TempDir()
			err = os.WriteFile(filepath.Join(path, "main.tf"), []byte(config), 0o777)
			require.NoError(t, err)
			tarball, err := internal.Pack(path)
			require.NoError(t, err)
			err = daemon.UploadConfig(ctx, cv.ID, tarball)
			require.NoError(t, err)

			// watch run events
			runsSub, runsUnsub := daemon.WatchRuns()
			defer runsUnsub()

			// watch log events
			logsSub, logsUnsub := daemon.WatchLogs()
			defer logsUnsub()

			// create run
			_ = daemon.createRun(t, ctx, ws, cv)

			// wait for the run to report an error status and for the logs to contain
			// the error message.
			var (
				gotErrorStatus bool
				gotErrorLogs   bool
			)
			errorRegex := regexp.MustCompile(`Error: exit status 1: Error: Invalid resource type on main.tf line 5, in resource "null_resourc" "e2e": 5: resource "null_resourc" "e2e" {} The provider hashicorp/null does not support resource type "null_resourc". Did you mean "null_resource"?`)
			require.NoError(t, err)
			for {
				select {
				case event := <-logsSub:
					if errorRegex.Match(event.Payload.Data) {
						gotErrorLogs = true
					}
				case event := <-runsSub:
					if event.Payload.Status == run.RunErrored {
						gotErrorStatus = true
					}
				}
				if gotErrorLogs && gotErrorStatus {
					return
				}
			}
		})
	}
}
