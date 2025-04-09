package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_schedule(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wsID := resource.NewTfeID(resource.WorkspaceKind)

	t.Run("enqueue plan for pending plan-only run but don't add to workspace queue", func(t *testing.T) {
		run := &Run{Status: runstatus.Pending, PlanOnly: true, ID: resource.NewTfeID(resource.RunKind)}
		runClient := &fakeSchedulerRunClient{}
		s := &scheduler{
			runs:       runClient,
			workspaces: &fakeSchedulerWorkspaceClient{},
			queues:     make(map[resource.TfeID]queue),
		}

		err := s.schedule(ctx, wsID, run)
		require.NoError(t, err)

		assert.Equal(t, run.ID, *runClient.enqueuedRunID)
		assert.Equal(t, 0, len(s.queues))
	})

	t.Run("ignore running plan-only run", func(t *testing.T) {
		run := &Run{Status: runstatus.Planning, PlanOnly: true, ID: resource.NewTfeID(resource.RunKind)}
		runClient := &fakeSchedulerRunClient{}
		s := &scheduler{
			runs:       runClient,
			workspaces: &fakeSchedulerWorkspaceClient{},
			queues:     make(map[resource.TfeID]queue),
		}

		err := s.schedule(ctx, wsID, run)
		require.NoError(t, err)

		assert.Nil(t, runClient.enqueuedRunID)
		assert.Equal(t, 0, len(s.queues))
	})

	t.Run("add several runs to queue and make oldest run current", func(t *testing.T) {
		s := &scheduler{
			runs:       &fakeSchedulerRunClient{},
			workspaces: &fakeSchedulerWorkspaceClient{},
			queues:     make(map[resource.TfeID]queue),
		}
		run1 := &Run{Status: runstatus.Pending, ID: resource.NewTfeID(resource.RunKind)}
		run2 := &Run{Status: runstatus.Pending, ID: resource.NewTfeID(resource.RunKind)}
		run3 := &Run{Status: runstatus.Pending, ID: resource.NewTfeID(resource.RunKind)}

		err := s.schedule(ctx, wsID, run1)
		require.NoError(t, err)
		err = s.schedule(ctx, wsID, run2)
		require.NoError(t, err)
		err = s.schedule(ctx, wsID, run3)
		require.NoError(t, err)

		assert.Equal(t, 1, len(s.queues))
		assert.Equal(t, run1.ID, *s.queues[wsID].current)
		assert.Equal(t, []resource.TfeID{run2.ID, run3.ID}, s.queues[wsID].backlog)
	})

	t.Run("attempt to schedule run on user-locked workspace", func(t *testing.T) {
		s := &scheduler{
			runs: &fakeSchedulerRunClient{
				enqueuePlanError: workspace.ErrWorkspaceAlreadyLocked,
			},
			workspaces: &fakeSchedulerWorkspaceClient{},
			queues:     make(map[resource.TfeID]queue),
		}
		run1 := &Run{Status: runstatus.Pending, ID: resource.NewTfeID(resource.RunKind)}

		// Should not propagate error
		err := s.schedule(ctx, wsID, run1)
		assert.NoError(t, err)
		// Should not be made current run but instead placed on backlog
		assert.Equal(t, 1, len(s.queues))
		assert.Nil(t, s.queues[wsID].current)
		assert.Equal(t, []resource.TfeID{run1.ID}, s.queues[wsID].backlog)
	})

	t.Run("remove finished current run and unlock queue", func(t *testing.T) {
		runID := resource.NewTfeID(resource.RunKind)
		workspaces := &fakeSchedulerWorkspaceClient{}
		s := &scheduler{
			runs:       &fakeSchedulerRunClient{},
			workspaces: workspaces,
			queues: map[resource.TfeID]queue{
				wsID: {current: &runID},
			},
		}

		err := s.schedule(ctx, wsID, &Run{Status: runstatus.Applied, ID: runID})
		require.NoError(t, err)

		assert.Nil(t, s.queues[wsID].current)
		assert.True(t, workspaces.unlocked)
	})
}

func TestScheduler_process(t *testing.T) {
	runID1 := resource.NewTfeID(resource.RunKind)
	runID2 := resource.NewTfeID(resource.RunKind)

	tests := []struct {
		name        string
		q           queue
		run         *Run
		want        queue
		enqueuePlan bool
		unlock      bool
	}{
		{
			"do nothing with empty queue",
			queue{},
			nil,
			queue{},
			false,
			false,
		},
		{
			"make pending run current and request enqueue plan",
			queue{},
			&Run{ID: runID1, Status: runstatus.Pending},
			queue{current: &runID1},
			true,
			false,
		},
		{
			"make plan_enqueued run current and do not request enqueue plan",
			queue{},
			&Run{ID: runID1, Status: runstatus.PlanQueued},
			queue{current: &runID1},
			false,
			false,
		},
		{
			"move backlogged run into current and request enqueue plan",
			queue{backlog: []resource.TfeID{runID1}},
			nil,
			queue{current: &runID1},
			true,
			false,
		},
		{
			"remove finished run from current and unlock queue",
			queue{current: &runID1},
			&Run{ID: runID1, Status: runstatus.Applied},
			queue{},
			false,
			true,
		},
		{
			"remove finished run from backlog",
			queue{backlog: []resource.TfeID{runID1}},
			&Run{ID: runID1, Status: runstatus.Applied},
			queue{},
			false,
			false,
		},
		{
			"remove finished run from current and make backlogged run current and request enqueue plan",
			queue{current: &runID1, backlog: []resource.TfeID{runID2}},
			&Run{ID: runID1, Status: runstatus.Applied},
			queue{current: &runID2},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, enqueuePlan, unlock := tt.q.process(tt.run)

			assert.Equal(t, tt.want.current, got.current)
			assert.Equal(t, len(tt.want.backlog), len(got.backlog))
			for i := range got.backlog {
				assert.Equal(t, tt.want.backlog[i], got.backlog[i])
			}
			assert.Equal(t, tt.enqueuePlan, enqueuePlan)
			assert.Equal(t, tt.unlock, unlock)
		})
	}
}

type fakeSchedulerRunClient struct {
	schedulerRunClient

	enqueuedRunID    *resource.TfeID
	enqueuePlanError error
}

func (f *fakeSchedulerRunClient) EnqueuePlan(ctx context.Context, runID resource.TfeID) (*Run, error) {
	f.enqueuedRunID = &runID
	return nil, f.enqueuePlanError
}

type fakeSchedulerWorkspaceClient struct {
	schedulerWorkspaceClient

	unlocked bool
}

func (f *fakeSchedulerWorkspaceClient) Unlock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID, force bool) (*workspace.Workspace, error) {
	f.unlocked = true
	return nil, nil
}
