package agent

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllocator(t *testing.T) {
	tests := []struct {
		name string
		// seed allocator with pools
		pools []*Pool
		// seed allocator with agents
		agents []*Agent
		// seed allocator with job
		job *Job
		// want this job after allocation
		wantJob *Job
		// want these capacities after allocation
		wantCapacities map[string]int
	}{
		{
			name: "allocate unallocated job to idle agent",
			agents: []*Agent{
				{ID: "agent-idle", Status: AgentIdle, Concurrency: 1},
			},
			job: &Job{
				JobSpec:       JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.RemoteExecutionMode,
			},
			wantJob: &Job{
				JobSpec:       JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobAllocated,
				ExecutionMode: workspace.RemoteExecutionMode,
				AgentID:       internal.String("agent-idle"),
			},
			wantCapacities: map[string]int{
				"agent-idle": 0,
			},
		},
		{
			name: "allocate unallocated job to idle agent that has pinged more recently than another",
			agents: []*Agent{
				{ID: "agent-idle-most-recent", Status: AgentIdle, Concurrency: 1, LastPingAt: time.Now().Add(-time.Second)},
				{ID: "agent-idle-less-recent", Status: AgentIdle, Concurrency: 1, LastPingAt: time.Now()},
			},
			job: &Job{
				JobSpec:       JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.RemoteExecutionMode,
			},
			wantJob: &Job{
				JobSpec:       JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobAllocated,
				ExecutionMode: workspace.RemoteExecutionMode,
				AgentID:       internal.String("agent-idle-most-recent"),
			},
			wantCapacities: map[string]int{
				"agent-idle-most-recent": 0,
				"agent-idle-less-recent": 1,
			},
		},
		{
			name: "re-allocate job from unresponsive agent",
			agents: []*Agent{
				{ID: "agent-unknown", Status: AgentUnknown, Concurrency: 0},
				{ID: "agent-idle", Status: AgentIdle, Concurrency: 1},
			},
			job: &Job{
				JobSpec:       JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobAllocated,
				ExecutionMode: workspace.RemoteExecutionMode,
				AgentID:       internal.String("agent-unknown"),
			},
			wantJob: &Job{
				JobSpec:       JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobAllocated,
				ExecutionMode: workspace.RemoteExecutionMode,
				AgentID:       internal.String("agent-idle"),
			},
			wantCapacities: map[string]int{
				"agent-unknown": 1,
				"agent-idle":    0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &allocator{
				Service: &fakeService{
					job: tt.job,
				},
			}
			a.seed(tt.pools, tt.agents, []*Job{tt.job})
			err := a.allocate(context.Background())
			require.NoError(t, err)
			// check capacities
			assert.Equal(t, len(tt.wantCapacities), len(a.capacities))
			// check job
			assert.Equal(t, tt.wantJob, a.jobs[tt.wantJob.JobSpec])
		})
	}
}
