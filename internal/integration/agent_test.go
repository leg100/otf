package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Agents demonstrates the use of pooled agents
func TestIntegration_Agents(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)

	pool1, err := daemon.Runners.CreateAgentPool(ctx, runner.CreateAgentPoolOptions{
		Name:         "pool-1",
		Organization: org.Name,
	})
	require.NoError(t, err)

	pool2, err := daemon.Runners.CreateAgentPool(ctx, runner.CreateAgentPoolOptions{
		Name:         "pool-2",
		Organization: org.Name,
	})
	require.NoError(t, err)

	// ws1 is assigned to pool1
	ws1, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
		Name:          new("ws-1"),
		Organization:  &org.Name,
		ExecutionMode: internal.Ptr(workspace.AgentExecutionMode),
		AgentPoolID:   &pool1.ID,
	})
	require.NoError(t, err)

	// ws2 to assigned to pool2
	ws2, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
		Name:          new("ws-2"),
		Organization:  &org.Name,
		ExecutionMode: internal.Ptr(workspace.AgentExecutionMode),
		AgentPoolID:   &pool2.ID,
	})
	require.NoError(t, err)

	// start agents up
	agent1, shutdown1 := daemon.startAgent(t, ctx, org.Name, &pool1.ID, "")
	defer shutdown1()
	agent2, shutdown2 := daemon.startAgent(t, ctx, org.Name, &pool2.ID, "")
	defer shutdown2()

	// watch job events
	jobsSub, unsub := daemon.Runners.WatchJobs(ctx)
	defer unsub()

	// create a run on ws1
	_ = daemon.createRun(t, ctx, ws1, nil, nil)

	// wait for job to be allocated to agent1
	wait(t, jobsSub, func(event pubsub.Event[*runner.JobEvent]) bool {
		return event.Payload.Status == runner.JobAllocated &&
			*event.Payload.RunnerID == agent1.ID
	})

	// create a run on ws2
	_ = daemon.createRun(t, ctx, ws2, nil, nil)

	// wait for job to be allocated to agent2
	wait(t, jobsSub, func(event pubsub.Event[*runner.JobEvent]) bool {
		return event.Payload.Status == runner.JobAllocated &&
			*event.Payload.RunnerID == agent2.ID
	})
}
