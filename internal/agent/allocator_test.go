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

func TestAllocator_seed(t *testing.T) {
	pool1 := &Pool{ID: "pool-1"}
	pool2 := &Pool{ID: "pool-2"}

	agent1 := &Agent{ID: "agent-1", Status: AgentIdle, MaxJobs: 5}
	agent2 := &Agent{ID: "agent-2", Status: AgentIdle, MaxJobs: 5}

	job1 := &Job{
		Spec:   JobSpec{RunID: "run-1", Phase: internal.PlanPhase},
		Status: JobUnallocated,
	}
	job2 := &Job{
		Spec:    JobSpec{RunID: "run-2", Phase: internal.PlanPhase},
		Status:  JobAllocated,
		AgentID: internal.String("agent-2"),
	}

	a := &allocator{}
	a.seed([]*Pool{pool1, pool2}, []*Agent{agent1, agent2}, []*Job{job1, job2})

	if assert.Len(t, a.pools, 2) {
		assert.Contains(t, a.pools, "pool-1")
		assert.Contains(t, a.pools, "pool-2")
	}
	if assert.Len(t, a.agents, 2) {
		assert.Contains(t, a.agents, "agent-1")
		assert.Contains(t, a.agents, "agent-2")
	}
	if assert.Len(t, a.capacities, 2) {
		if assert.Contains(t, a.capacities, "agent-1") {
			assert.Equal(t, a.capacities["agent-1"], 5)
		}
		if assert.Contains(t, a.capacities, "agent-2") {
			assert.Equal(t, a.capacities["agent-2"], 4)
		}
	}
}

func TestAllocator_allocate(t *testing.T) {
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
			name: "allocate job to agent",
			agents: []*Agent{
				{ID: "agent-idle", Status: AgentIdle, MaxJobs: 1},
			},
			job: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.RemoteExecutionMode,
			},
			wantJob: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobAllocated,
				ExecutionMode: workspace.RemoteExecutionMode,
				AgentID:       internal.String("agent-idle"),
			},
			wantCapacities: map[string]int{
				"agent-idle": 0,
			},
		},
		{
			name: "allocate job to agent that has pinged more recently than another",
			agents: []*Agent{
				{ID: "agent-new", Status: AgentIdle, MaxJobs: 1, LastPingAt: time.Now()},
				{ID: "agent-old", Status: AgentIdle, MaxJobs: 1, LastPingAt: time.Now().Add(-time.Second)},
			},
			job: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.RemoteExecutionMode,
			},
			wantJob: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobAllocated,
				ExecutionMode: workspace.RemoteExecutionMode,
				AgentID:       internal.String("agent-new"),
			},
			wantCapacities: map[string]int{
				"agent-new": 0,
				"agent-old": 1,
			},
		},
		{
			name: "allocate job to organization-scoped pool agent",
			pools: []*Pool{
				{ID: "pool-1", OrganizationScoped: true, Organization: "acme-corp"},
			},
			agents: []*Agent{
				{ID: "agent-1", Status: AgentIdle, MaxJobs: 1, AgentPoolID: internal.String("pool-1")},
			},
			job: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.AgentExecutionMode,
				Organization:  "acme-corp",
			},
			wantJob: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobAllocated,
				ExecutionMode: workspace.AgentExecutionMode,
				AgentID:       internal.String("agent-1"),
				Organization:  "acme-corp",
			},
			wantCapacities: map[string]int{
				"agent-1": 0,
			},
		},
		{
			name: "allocate job to pool agent with matching assigned workspace",
			pools: []*Pool{
				{ID: "pool-1", AssignedWorkspaces: []string{"workspace-1"}},
			},
			agents: []*Agent{
				{ID: "agent-1", Status: AgentIdle, MaxJobs: 1, AgentPoolID: internal.String("pool-1")},
			},
			job: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.AgentExecutionMode,
				WorkspaceID:   "workspace-1",
			},
			wantJob: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobAllocated,
				ExecutionMode: workspace.AgentExecutionMode,
				AgentID:       internal.String("agent-1"),
				WorkspaceID:   "workspace-1",
			},
			wantCapacities: map[string]int{
				"agent-1": 0,
			},
		},
		{
			name: "do not allocate job to pool agent in different org",
			pools: []*Pool{
				{ID: "pool-1", OrganizationScoped: true, Organization: "acme-corp"},
			},
			agents: []*Agent{
				{ID: "agent-1", Status: AgentIdle, MaxJobs: 1, AgentPoolID: internal.String("pool-1")},
			},
			job: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.AgentExecutionMode,
				Organization:  "enron",
			},
			wantJob: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.AgentExecutionMode,
				Organization:  "enron",
			},
			wantCapacities: map[string]int{
				"agent-1": 1,
			},
		},
		{
			name: "do not allocate job to pool agent with non-matching assigned workspace",
			pools: []*Pool{
				{ID: "pool-1", AssignedWorkspaces: []string{"workspace-non-matching"}},
			},
			agents: []*Agent{
				{ID: "agent-1", Status: AgentIdle, MaxJobs: 1, AgentPoolID: internal.String("pool-1")},
			},
			job: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.AgentExecutionMode,
				WorkspaceID:   "workspace-1",
			},
			wantJob: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobUnallocated,
				ExecutionMode: workspace.AgentExecutionMode,
				WorkspaceID:   "workspace-1",
			},
			wantCapacities: map[string]int{
				"agent-1": 1,
			},
		},
		{
			name: "re-allocate job from unresponsive agent",
			agents: []*Agent{
				{ID: "agent-unknown", Status: AgentUnknown, MaxJobs: 0},
				{ID: "agent-idle", Status: AgentIdle, MaxJobs: 1},
			},
			job: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:        JobAllocated,
				ExecutionMode: workspace.RemoteExecutionMode,
				AgentID:       internal.String("agent-unknown"),
			},
			wantJob: &Job{
				Spec:          JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
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
			assert.Equal(t, tt.wantJob, a.jobs[tt.wantJob.Spec])
		})
	}
}
