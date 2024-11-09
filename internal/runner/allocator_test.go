package runner

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllocator_seed(t *testing.T) {
	agent1 := &RunnerMeta{ID: resource.NewID(resource.RunnerKind), Status: RunnerIdle, MaxJobs: 5}
	agent2 := &RunnerMeta{ID: resource.NewID(resource.RunnerKind), Status: RunnerIdle, MaxJobs: 5}

	job1 := &Job{
		RunID:  resource.NewID(resource.RunKind),
		Phase:  internal.PlanPhase,
		Status: JobUnallocated,
	}
	job2 := &Job{
		RunID:    resource.NewID(resource.RunKind),
		Phase:    internal.PlanPhase,
		Status:   JobAllocated,
		RunnerID: &agent2.ID,
	}

	a := &allocator{}
	a.seed([]*RunnerMeta{agent1, agent2}, []*Job{job1, job2})

	if assert.Len(t, a.runners, 2) {
		assert.Contains(t, a.runners, agent1.ID)
		assert.Contains(t, a.runners, agent2.ID)
	}
}

func TestAllocator_allocate(t *testing.T) {
	now := internal.CurrentTimestamp(nil)

	runner1ID := resource.NewID(resource.RunnerKind)
	runner2ID := resource.NewID(resource.RunnerKind)
	pool1ID := resource.NewID(resource.AgentPoolKind)
	job1ID := resource.NewID(resource.JobKind)

	tests := []struct {
		name string
		// seed allocator with pools
		pools []*Pool
		// seed allocator with runners
		runners []*RunnerMeta
		// seed allocator with job
		job *Job
		// want this job after allocation
		wantJob *Job
		// want these runners after allocation
		wantRunners map[resource.ID]*RunnerMeta
	}{
		{
			name: "allocate job to server runner",
			runners: []*RunnerMeta{
				{ID: runner1ID, Status: RunnerIdle, MaxJobs: 1},
			},
			job: &Job{
				RunID:  testutils.ParseID(t, "run-123"),
				Phase:  internal.PlanPhase,
				Status: JobUnallocated,
			},
			wantJob: &Job{
				RunID:    testutils.ParseID(t, "run-123"),
				Phase:    internal.PlanPhase,
				Status:   JobAllocated,
				RunnerID: &runner1ID,
			},
			wantRunners: map[resource.ID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1},
			},
		},
		{
			name: "allocate job to agent that has pinged more recently than another",
			runners: []*RunnerMeta{
				{ID: runner1ID, Status: RunnerIdle, MaxJobs: 1, LastPingAt: now},
				{ID: runner2ID, Status: RunnerIdle, MaxJobs: 1, LastPingAt: now.Add(-time.Second)},
			},
			job: &Job{
				ID:     job1ID,
				Status: JobUnallocated,
			},
			wantJob: &Job{
				ID:       job1ID,
				Status:   JobAllocated,
				RunnerID: &runner1ID,
			},
			wantRunners: map[resource.ID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1, LastPingAt: now},
				runner2ID: {ID: runner2ID, Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 0, LastPingAt: now.Add(-time.Second)},
			},
		},
		{
			name:  "allocate job to pool agent",
			pools: []*Pool{{ID: testutils.ParseID(t, "pool-1")}},
			runners: []*RunnerMeta{
				{
					ID:        runner1ID,
					Status:    RunnerIdle,
					MaxJobs:   1,
					AgentPool: &RunnerMetaAgentPool{ID: pool1ID},
				},
			},
			job: &Job{
				ID:          job1ID,
				Phase:       internal.PlanPhase,
				Status:      JobUnallocated,
				AgentPoolID: &pool1ID,
			},
			wantJob: &Job{
				ID:          job1ID,
				Phase:       internal.PlanPhase,
				Status:      JobAllocated,
				AgentPoolID: &pool1ID,
				RunnerID:    &runner1ID,
			},
			wantRunners: map[resource.ID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1, AgentPool: &RunnerMetaAgentPool{ID: pool1ID}},
			},
		},
		{
			name:  "do not allocate job to agent with insufficient capacity",
			pools: []*Pool{{ID: pool1ID}},
			runners: []*RunnerMeta{
				{ID: runner1ID, Status: RunnerIdle, CurrentJobs: 1, MaxJobs: 1},
			},
			job: &Job{
				ID:     job1ID,
				Status: JobUnallocated,
			},
			wantJob: &Job{
				ID:     job1ID,
				Status: JobUnallocated,
			},
			wantRunners: map[resource.ID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1},
			},
		},
		{
			name: "re-allocate job from unresponsive agent",
			runners: []*RunnerMeta{
				{ID: runner1ID, Status: RunnerUnknown, CurrentJobs: 1},
				{ID: runner2ID, Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 0},
			},
			job: &Job{
				ID:       job1ID,
				Status:   JobAllocated,
				RunnerID: &runner1ID,
			},
			wantJob: &Job{
				ID:       job1ID,
				Status:   JobAllocated,
				RunnerID: &runner2ID,
			},
			wantRunners: map[resource.ID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerUnknown, CurrentJobs: 0},
				runner2ID: {ID: runner2ID, Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1},
			},
		},
		{
			name:    "de-allocate finished job",
			runners: []*RunnerMeta{{ID: runner1ID, CurrentJobs: 1}},
			job: &Job{
				ID:       job1ID,
				Status:   JobFinished,
				RunnerID: &runner1ID,
			},
			wantJob:     nil,
			wantRunners: map[resource.ID]*RunnerMeta{runner1ID: {ID: runner1ID, CurrentJobs: 0}},
		},
		{
			name: "ignore running job",
			job: &Job{
				ID:       job1ID,
				Status:   JobRunning,
				RunnerID: &runner1ID,
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
			a.seed(tt.runners, []*Job{tt.job})
			err := a.allocate(context.Background())
			require.NoError(t, err)
			// check agents
			if assert.Equal(t, len(tt.wantRunners), len(a.runners)) {
				for id, want := range tt.wantRunners {
					assert.Equal(t, want, a.runners[id])
				}
			}
			// check job
			if tt.wantJob != nil {
				assert.Equal(t, tt.wantJob, a.jobs[tt.wantJob.ID])
			}
		})
	}
}
