package runner

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllocator_seed(t *testing.T) {
	pool1 := &Pool{ID: "pool-1"}
	pool2 := &Pool{ID: "pool-2"}

	agent1 := &runnerMeta{ID: "agent-1", Status: RunnerIdle, MaxJobs: 5}
	agent2 := &runnerMeta{ID: "agent-2", Status: RunnerIdle, MaxJobs: 5}

	job1 := &Job{
		Spec:   JobSpec{RunID: "run-1", Phase: internal.PlanPhase},
		Status: JobUnallocated,
	}
	job2 := &Job{
		Spec:     JobSpec{RunID: "run-2", Phase: internal.PlanPhase},
		Status:   JobAllocated,
		RunnerID: internal.String("agent-2"),
	}

	a := &allocator{}
	a.seed([]*Pool{pool1, pool2}, []*runnerMeta{agent1, agent2}, []*Job{job1, job2})

	if assert.Len(t, a.pools, 2) {
		assert.Contains(t, a.pools, "pool-1")
		assert.Contains(t, a.pools, "pool-2")
	}
	if assert.Len(t, a.runners, 2) {
		assert.Contains(t, a.runners, "agent-1")
		assert.Contains(t, a.runners, "agent-2")
	}
}

func TestAllocator_allocate(t *testing.T) {
	now := internal.CurrentTimestamp(nil)

	tests := []struct {
		name string
		// seed allocator with pools
		pools []*Pool
		// seed allocator with agents
		agents []*runnerMeta
		// seed allocator with job
		job *Job
		// want this job after allocation
		wantJob *Job
		// want these agents after allocation
		wantAgents map[string]*runnerMeta
	}{
		{
			name: "allocate job to server agent",
			agents: []*runnerMeta{
				{ID: "agent-idle", Status: RunnerIdle, MaxJobs: 1},
			},
			job: &Job{
				Spec:   JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status: JobUnallocated,
			},
			wantJob: &Job{
				Spec:     JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:   JobAllocated,
				RunnerID: internal.String("agent-idle"),
			},
			wantAgents: map[string]*runnerMeta{
				"agent-idle": {ID: "agent-idle", Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1},
			},
		},
		{
			name: "allocate job to agent that has pinged more recently than another",
			agents: []*runnerMeta{
				{ID: "agent-new", Status: RunnerIdle, MaxJobs: 1, LastPingAt: now},
				{ID: "agent-old", Status: RunnerIdle, MaxJobs: 1, LastPingAt: now.Add(-time.Second)},
			},
			job: &Job{
				Spec:   JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status: JobUnallocated,
			},
			wantJob: &Job{
				Spec:     JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:   JobAllocated,
				RunnerID: internal.String("agent-new"),
			},
			wantAgents: map[string]*runnerMeta{
				"agent-new": {ID: "agent-new", Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1, LastPingAt: now},
				"agent-old": {ID: "agent-old", Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 0, LastPingAt: now.Add(-time.Second)},
			},
		},
		{
			name:  "allocate job to pool agent",
			pools: []*Pool{{ID: "pool-1"}},
			agents: []*runnerMeta{
				{ID: "agent-1", Status: RunnerIdle, MaxJobs: 1, AgentPoolID: internal.String("pool-1")},
			},
			job: &Job{
				Spec:        JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:      JobUnallocated,
				AgentPoolID: internal.String("pool-1"),
			},
			wantJob: &Job{
				Spec:        JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:      JobAllocated,
				AgentPoolID: internal.String("pool-1"),
				RunnerID:    internal.String("agent-1"),
			},
			wantAgents: map[string]*runnerMeta{
				"agent-1": {ID: "agent-1", Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1, AgentPoolID: internal.String("pool-1")},
			},
		},
		{
			name:  "do not allocate job to agent with insufficient capacity",
			pools: []*Pool{{ID: "pool-1"}},
			agents: []*runnerMeta{
				{ID: "agent-1", Status: RunnerIdle, CurrentJobs: 1, MaxJobs: 1},
			},
			job: &Job{
				Spec:   JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status: JobUnallocated,
			},
			wantJob: &Job{
				Spec:   JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status: JobUnallocated,
			},
			wantAgents: map[string]*runnerMeta{
				"agent-1": {ID: "agent-1", Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1},
			},
		},
		{
			name: "re-allocate job from unresponsive agent",
			agents: []*runnerMeta{
				{ID: "agent-unknown", Status: RunnerUnknown, CurrentJobs: 1},
				{ID: "agent-idle", Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 0},
			},
			job: &Job{
				Spec:     JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:   JobAllocated,
				RunnerID: internal.String("agent-unknown"),
			},
			wantJob: &Job{
				Spec:     JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:   JobAllocated,
				RunnerID: internal.String("agent-idle"),
			},
			wantAgents: map[string]*runnerMeta{
				"agent-unknown": {ID: "agent-unknown", Status: RunnerUnknown, CurrentJobs: 0},
				"agent-idle":    {ID: "agent-idle", Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1},
			},
		},
		{
			name:   "de-allocate finished job",
			agents: []*runnerMeta{{ID: "agent-1", CurrentJobs: 1}},
			job: &Job{
				Spec:     JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:   JobFinished,
				RunnerID: internal.String("agent-1"),
			},
			wantJob:    nil,
			wantAgents: map[string]*runnerMeta{"agent-1": {ID: "agent-1", CurrentJobs: 0}},
		},
		{
			name: "ignore running job",
			job: &Job{
				Spec:     JobSpec{RunID: "run-123", Phase: internal.PlanPhase},
				Status:   JobRunning,
				RunnerID: internal.String("agent-1"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &allocator{
				Logger: logr.Discard(),
				client: &fakeService{
					job: tt.job,
				},
			}
			a.seed(tt.pools, tt.agents, []*Job{tt.job})
			err := a.allocate(context.Background())
			require.NoError(t, err)
			// check agents
			if assert.Equal(t, len(tt.wantAgents), len(a.runners)) {
				for id, want := range tt.wantAgents {
					assert.Equal(t, want, a.runners[id])
				}
			}
			// check job
			if tt.wantJob != nil {
				assert.Equal(t, tt.wantJob, a.jobs[tt.wantJob.Spec])
			}
		})
	}
}
