package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

// TestLocal demonstrates usage of the local execution mode, whereby OTF is only
// used as remote state storage.
func TestLocal(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	// create workspace with local execution mode
	_, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
		Name:          internal.String("local-ws"),
		Organization:  internal.String(org.Name),
		ExecutionMode: workspace.ExecutionModePtr(workspace.LocalExecutionMode),
	})
	require.NoError(t, err)

	// create root module, setting otfd1 as hostname
	root := newRootModule(t, daemon.System.Hostname(), org.Name, "local-ws")

	// run terraform locally, configuring OTF as a remote backend.
	daemon.tfcli(t, ctx, "init", root)
	out := daemon.tfcli(t, ctx, "plan", root)
	require.Contains(t, out, "Plan: 1 to add, 0 to change, 0 to destroy.")
	out = daemon.tfcli(t, ctx, "apply", root, "-auto-approve")
	require.Contains(t, out, "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
}
