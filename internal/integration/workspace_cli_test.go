package integration

import (
	"encoding/json"
	"testing"

	"github.com/leg100/otf/internal/agent"
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
	var (
		ws1, ws2, ws3 workspace.Workspace
	)
	ws1out := daemon.otfcli(t, ctx, "workspaces", "new", "--name", "ws1", "--organization", org.Name)
	err := json.Unmarshal([]byte(ws1out), &ws1)
	require.NoError(t, err)

	ws2out := daemon.otfcli(t, ctx, "workspaces", "new", "--name", "ws2", "--organization", org.Name)
	err = json.Unmarshal([]byte(ws2out), &ws1)
	require.NoError(t, err)

	ws3out := daemon.otfcli(t, ctx, "workspaces", "new", "--name", "ws3", "--organization", org.Name)
	err = json.Unmarshal([]byte(ws3out), &ws1)
	require.NoError(t, err)

	// list workspaces via CLI
	out := daemon.otfcli(t, ctx, "workspaces", "list", "--organization", org.Name)
	assert.Contains(t, out, ws1.Name)
	assert.Contains(t, out, ws2.Name)
	assert.Contains(t, out, ws3.Name)

	// show workspace via CLI (outputs as JSON)
	out = daemon.otfcli(t, ctx, "workspaces", "show", "--organization", org.Name, ws1.Name)
	var got workspace.Workspace
	err = json.Unmarshal([]byte(out), &got)
	require.NoError(t, err)

	// edit workspace via CLI
	//
	// create pool first so that one can be specified in the CLI command
	pool, err := daemon.Agents.CreateAgentPool(ctx, agent.CreateAgentPoolOptions{
		Organization: org.Name,
		Name:         "pool-1",
	})
	require.NoError(t, err)
	out = daemon.otfcli(t, ctx, "workspaces", "edit", "--organization", org.Name,
		ws1.Name, "--execution-mode", "agent", "--agent-pool-id", pool.ID)
	assert.Equal(t, "updated workspace\n", out)
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
