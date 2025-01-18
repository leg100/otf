package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_schedule(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wsID := resource.NewID(resource.WorkspaceKind)

	t.Run("dont add plan-only runs to queues", func(t *testing.T) {
		s := &scheduler{
			runs:       &fakeSchedulerRunClient{},
			workspaces: &fakeSchedulerWorkspaceClient{},
			queues:     make(map[resource.ID]queue),
		}

		run := &Run{Status: RunPending, PlanOnly: true}
		err := s.schedule(ctx, wsID, run)
		require.NoError(t, err)

		assert.Equal(t, 0, len(s.queues))
	})

	t.Run("add several runs to queue and make oldest run current", func(t *testing.T) {
		s := &scheduler{
			runs:       &fakeSchedulerRunClient{},
			workspaces: &fakeSchedulerWorkspaceClient{},
			queues:     make(map[resource.ID]queue),
		}
		run1 := &Run{Status: RunPending, ID: resource.NewID(resource.RunKind)}
		run2 := &Run{Status: RunPending, ID: resource.NewID(resource.RunKind)}
		run3 := &Run{Status: RunPending, ID: resource.NewID(resource.RunKind)}

		err := s.schedule(ctx, wsID, run1)
		require.NoError(t, err)
		err = s.schedule(ctx, wsID, run2)
		require.NoError(t, err)
		err = s.schedule(ctx, wsID, run3)
		require.NoError(t, err)

		assert.Equal(t, 1, len(s.queues))
		assert.Equal(t, run1.ID, *s.queues[wsID].current)
		assert.Equal(t, []resource.ID{run2.ID, run3.ID}, s.queues[wsID].backlog)
	})

	t.Run("remove finished current run", func(t *testing.T) {
		runID := resource.NewID(resource.RunKind)
		s := &scheduler{
			runs:       &fakeSchedulerRunClient{},
			workspaces: &fakeSchedulerWorkspaceClient{},
			queues: map[resource.ID]queue{
				wsID: {current: &runID},
			},
		}

		err := s.schedule(ctx, wsID, &Run{Status: RunApplied, ID: runID})
		require.NoError(t, err)

		assert.Nil(t, s.queues[wsID].current)
	})
}

func TestScheduler_process(t *testing.T) {
	runID1 := resource.NewID(resource.RunKind)
	runID2 := resource.NewID(resource.RunKind)

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
			"make run current and request enqueue plan",
			queue{},
			&Run{ID: runID1},
			queue{current: &runID1},
			true,
			false,
		},
		{
			"move backlogged run into current and request enqueue plan",
			queue{backlog: []resource.ID{runID1}},
			nil,
			queue{current: &runID1},
			true,
			false,
		},
		{
			"remove finished run from current and unlock queue",
			queue{current: &runID1},
			&Run{ID: runID1, Status: RunApplied},
			queue{},
			false,
			true,
		},
		{
			"remove finished run from backlog",
			queue{backlog: []resource.ID{runID1}},
			&Run{ID: runID1, Status: RunApplied},
			queue{},
			false,
			false,
		},
		{
			"remove finished run from current and make backlogged run current and request enqueue plan",
			queue{current: &runID1, backlog: []resource.ID{runID2}},
			&Run{ID: runID1, Status: RunApplied},
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
}

func (f *fakeSchedulerRunClient) EnqueuePlan(ctx context.Context, runID resource.ID) (*Run, error) {
	return nil, nil
}

type fakeSchedulerWorkspaceClient struct {
	schedulerWorkspaceClient
}

func (f *fakeSchedulerWorkspaceClient) Unlock(ctx context.Context, workspaceID resource.ID, runID *resource.ID, force bool) (*workspace.Workspace, error) {
	return nil, nil
}
