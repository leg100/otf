package runner

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllocator_seed(t *testing.T) {
	agent1 := &RunnerMeta{ID: resource.NewTfeID(resource.RunnerKind), Status: RunnerIdle, MaxJobs: 5}
	agent2 := &RunnerMeta{ID: resource.NewTfeID(resource.RunnerKind), Status: RunnerIdle, MaxJobs: 5}

	job1 := &Job{
		RunID:  resource.NewTfeID(resource.RunKind),
		Phase:  run.PlanPhase,
		Status: JobUnallocated,
	}
	job2 := &Job{
		RunID:    resource.NewTfeID(resource.RunKind),
		Phase:    run.PlanPhase,
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

	runner1ID := resource.NewTfeID(resource.RunnerKind)
	runner2ID := resource.NewTfeID(resource.RunnerKind)
	pool1ID := resource.NewTfeID(resource.AgentPoolKind)
	job1ID := resource.NewTfeID(resource.JobKind)

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
		wantRunners map[resource.TfeID]*RunnerMeta
		// want this tally of current jobs after allocation
		wantCurrentJobs map[resource.TfeID]int
	}{
		{
			name: "allocate job to server runner",
			runners: []*RunnerMeta{
				{ID: runner1ID, Status: RunnerIdle, MaxJobs: 1},
			},
			job: &Job{
				RunID:  testutils.ParseID(t, "run-123"),
				Phase:  run.PlanPhase,
				Status: JobUnallocated,
			},
			wantJob: &Job{
				RunID:    testutils.ParseID(t, "run-123"),
				Phase:    run.PlanPhase,
				Status:   JobAllocated,
				RunnerID: &runner1ID,
			},
			wantRunners: map[resource.TfeID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerIdle, MaxJobs: 1},
			},
			wantCurrentJobs: map[resource.TfeID]int{runner1ID: 1},
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
			wantRunners: map[resource.TfeID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerIdle, MaxJobs: 1, LastPingAt: now},
				runner2ID: {ID: runner2ID, Status: RunnerIdle, MaxJobs: 1, LastPingAt: now.Add(-time.Second)},
			},
			wantCurrentJobs: map[resource.TfeID]int{runner1ID: 1},
		},
		{
			name:  "allocate job to pool agent",
			pools: []*Pool{{ID: testutils.ParseID(t, "pool-1")}},
			runners: []*RunnerMeta{
				{
					ID:        runner1ID,
					Status:    RunnerIdle,
					MaxJobs:   1,
					AgentPool: &Pool{ID: pool1ID},
				},
			},
			job: &Job{
				ID:          job1ID,
				Phase:       run.PlanPhase,
				Status:      JobUnallocated,
				AgentPoolID: &pool1ID,
			},
			wantJob: &Job{
				ID:          job1ID,
				Phase:       run.PlanPhase,
				Status:      JobAllocated,
				AgentPoolID: &pool1ID,
				RunnerID:    &runner1ID,
			},
			wantRunners: map[resource.TfeID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerIdle, MaxJobs: 1, AgentPool: &Pool{ID: pool1ID}},
			},
			wantCurrentJobs: map[resource.TfeID]int{runner1ID: 1},
		},
		{
			name:  "do not allocate job to agent with insufficient capacity",
			pools: []*Pool{{ID: pool1ID}},
			runners: []*RunnerMeta{
				{ID: runner1ID, Status: RunnerIdle, CurrentJobs: 1, MaxJobs: 1, ExecutorKind: ForkExecutorKind},
			},
			job: &Job{
				ID:     job1ID,
				Status: JobUnallocated,
			},
			wantJob: &Job{
				ID:     job1ID,
				Status: JobUnallocated,
			},
			wantRunners: map[resource.TfeID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerIdle, MaxJobs: 1, CurrentJobs: 1, ExecutorKind: ForkExecutorKind},
			},
		},
		{
			name: "re-allocate job from unresponsive agent",
			runners: []*RunnerMeta{
				{ID: runner1ID, Status: RunnerUnknown},
				{ID: runner2ID, Status: RunnerIdle, MaxJobs: 1},
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
			wantRunners: map[resource.TfeID]*RunnerMeta{
				runner1ID: {ID: runner1ID, Status: RunnerUnknown},
				runner2ID: {ID: runner2ID, Status: RunnerIdle, MaxJobs: 1},
			},
			wantCurrentJobs: map[resource.TfeID]int{runner2ID: 1},
		},
		{
			name:    "deallocate finished job",
			runners: []*RunnerMeta{{ID: runner1ID, CurrentJobs: 1}},
			job: &Job{
				ID:       job1ID,
				Status:   JobFinished,
				RunnerID: &runner1ID,
			},
			wantJob:         nil,
			wantRunners:     map[resource.TfeID]*RunnerMeta{runner1ID: {ID: runner1ID, CurrentJobs: 1}},
			wantCurrentJobs: map[resource.TfeID]int{runner1ID: 0},
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
			err := a.allocate(context.Background(), tt.job)
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
