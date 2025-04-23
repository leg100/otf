package integration

import (
	"encoding/json"
	"testing"

	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_WorkspaceCLI tests managing workspaces via the otf CLI
func TestIntegration_WorkspaceCLI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t)

	// create organization
	org := daemon.createOrganization(t, ctx)
	// create workspaces
	ws1 := daemon.createWorkspace(t, ctx, org)
	ws2 := daemon.createWorkspace(t, ctx, org)
	ws3 := daemon.createWorkspace(t, ctx, org)

	// list workspaces via CLI
	out := daemon.otfCLI(t, ctx, "workspaces", "list", "--organization", org.Name.String())
	assert.Contains(t, out, ws1.Name)
	assert.Contains(t, out, ws2.Name)
	assert.Contains(t, out, ws3.Name)

	// show workspace via CLI (outputs as JSON)
	out = daemon.otfCLI(t, ctx, "workspaces", "show", "--organization", org.Name.String(), ws1.Name)
	var got workspace.Workspace
	err := json.Unmarshal([]byte(out), &got)
	require.NoError(t, err)

	// edit workspace via CLI
	//
	// create pool first so that one can be specified in the CLI command
	pool, err := daemon.Runners.CreateAgentPool(ctx, runner.CreateAgentPoolOptions{
		Organization: org.Name,
		Name:         "pool-1",
	})
	require.NoError(t, err)
	out = daemon.otfCLI(t, ctx, "workspaces", "edit", "--organization", org.Name.String(),
		ws1.Name, "--execution-mode", "agent", "--agent-pool-id", pool.ID.String())
	assert.Equal(t, "updated workspace\n", out)
	assert.Equal(t, workspace.AgentExecutionMode, daemon.getWorkspace(t, ctx, ws1.ID).ExecutionMode)

	// lock/unlock/force-unlock workspace
	daemon.otfCLI(t, ctx, "workspaces", "lock", ws1.Name, "--organization", org.Name.String())
	assert.True(t, daemon.getWorkspace(t, ctx, ws1.ID).Locked())

	daemon.otfCLI(t, ctx, "workspaces", "unlock", ws1.Name, "--organization", org.Name.String())
	assert.False(t, daemon.getWorkspace(t, ctx, ws1.ID).Locked())

	daemon.otfCLI(t, ctx, "workspaces", "lock", ws1.Name, "--organization", org.Name.String())
	assert.True(t, daemon.getWorkspace(t, ctx, ws1.ID).Locked())

	daemon.otfCLI(t, ctx, "workspaces", "unlock", ws1.Name, "--organization", org.Name.String(), "--force")
	assert.False(t, daemon.getWorkspace(t, ctx, ws1.ID).Locked())
}
