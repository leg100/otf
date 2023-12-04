package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	agentpkg "github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Agents demonstrates the use of pooled agents
func TestIntegration_Agents(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	pool1, err := daemon.CreateAgentPool(ctx, agentpkg.CreateAgentPoolOptions{
		Name:         "pool-1",
		Organization: org.Name,
	})
	require.NoError(t, err)

	pool2, err := daemon.CreateAgentPool(ctx, agentpkg.CreateAgentPoolOptions{
		Name:         "pool-2",
		Organization: org.Name,
	})
	require.NoError(t, err)

	// ws1 is assigned to pool1
	ws1, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:          internal.String("ws-1"),
		Organization:  internal.String(org.Name),
		ExecutionMode: workspace.ExecutionModePtr(workspace.AgentExecutionMode),
		AgentPoolID:   internal.String(pool1.ID),
	})
	require.NoError(t, err)

	// ws2 to assigned to pool2
	ws2, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:          internal.String("ws-2"),
		Organization:  internal.String(org.Name),
		ExecutionMode: workspace.ExecutionModePtr(workspace.AgentExecutionMode),
		AgentPoolID:   internal.String(pool2.ID),
	})
	require.NoError(t, err)

	// start agents up
	agent1, shutdown1 := daemon.startAgent(t, ctx, org.Name, pool1.ID, "", agentpkg.Config{})
	defer shutdown1()
	agent2, shutdown2 := daemon.startAgent(t, ctx, org.Name, pool2.ID, "", agentpkg.Config{})
	defer shutdown2()

	// watch job events
	jobsSub, unsub := daemon.WatchJobs(ctx)
	defer unsub()

	// create a run on ws1
	_ = daemon.createRun(t, ctx, ws1, nil)

	// wait for job to be allocated to agent1
	testutils.Wait(t, jobsSub, func(event pubsub.Event[*agentpkg.Job]) bool {
		return event.Payload.Status == agentpkg.JobAllocated &&
			*event.Payload.AgentID == agent1.ID
	})

	// create a run on ws2
	_ = daemon.createRun(t, ctx, ws2, nil)

	// wait for job to be allocated to agent2
	testutils.Wait(t, jobsSub, func(event pubsub.Event[*agentpkg.Job]) bool {
		return event.Payload.Status == agentpkg.JobAllocated &&
			*event.Payload.AgentID == agent2.ID
	})
}
