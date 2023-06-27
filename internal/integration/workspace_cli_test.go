package integration

import (
	"encoding/json"
	"testing"

	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_WorkspaceCLI tests managing workspaces via the otf CLI
func TestIntegration_WorkspaceCLI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, nil)

	// create organization
	org := daemon.createOrganization(t, ctx)
	// create workspaces
	ws1 := daemon.createWorkspace(t, ctx, org)
	ws2 := daemon.createWorkspace(t, ctx, org)
	ws3 := daemon.createWorkspace(t, ctx, org)

	// list workspaces via CLI
	out := daemon.otfcli(t, ctx, "workspaces", "list", "--organization", org.Name)
	assert.Contains(t, out, ws1.Name)
	assert.Contains(t, out, ws2.Name)
	assert.Contains(t, out, ws3.Name)

	// show workspace via CLI (outputs as JSON)
	out = daemon.otfcli(t, ctx, "workspaces", "show", "--organization", org.Name, ws1.Name)
	var got workspace.Workspace
	err := json.Unmarshal([]byte(out), &got)
	require.NoError(t, err)

	// edit workspace via CLI
	out = daemon.otfcli(t, ctx, "workspaces", "edit", "--organization", org.Name,
		ws1.Name, "--execution-mode", "agent")
	assert.Equal(t, "updated execution mode: agent\n", out)
	assert.Equal(t, workspace.AgentExecutionMode, daemon.getWorkspace(t, ctx, ws1.ID).ExecutionMode)

	// lock/unlock/force-unlock workspace
	daemon.otfcli(t, ctx, "workspaces", "lock", ws1.Name, "--organization", org.Name)
	assert.True(t, daemon.getWorkspace(t, ctx, ws1.ID).Locked())

	daemon.otfcli(t, ctx, "workspaces", "unlock", ws1.Name, "--organization", org.Name)
	assert.False(t, daemon.getWorkspace(t, ctx, ws1.ID).Locked())

	daemon.otfcli(t, ctx, "workspaces", "lock", ws1.Name, "--organization", org.Name)
	assert.True(t, daemon.getWorkspace(t, ctx, ws1.ID).Locked())

	daemon.otfcli(t, ctx, "workspaces", "unlock", ws1.Name, "--organization", org.Name, "--force")
	assert.False(t, daemon.getWorkspace(t, ctx, ws1.ID).Locked())
}
