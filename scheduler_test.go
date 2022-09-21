package otf

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_HandleRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	speculative := NewTestRun(t, TestRunCreateOptions{Speculative: true})
	run1 := NewTestRun(t, TestRunCreateOptions{})
	run2 := NewTestRun(t, TestRunCreateOptions{})

	app := &fakeSchedulerApp{
		runs: []*Run{speculative, run1, run2},
	}
	scheduler := NewScheduler(logr.Discard(), app)
	errch := make(chan error)
	go func() {
		errch <- scheduler.Start(ctx)
	}()

	err := scheduler.handleRun(ctx, speculative)
	require.NoError(t, err)
	assert.Equal(t, 0, len(app.queue))

	err = scheduler.handleRun(ctx, run1)
	require.NoError(t, err)
	assert.Equal(t, 1, len(app.queue))
	assert.Equal(t, RunPlanQueued, app.queue[0].Status())

	err = scheduler.handleRun(ctx, run2)
	require.NoError(t, err)
	assert.Equal(t, 2, len(app.queue))
	assert.Equal(t, RunPending, app.queue[1].Status())

	cancel()
	assert.EqualError(t, <-errch, "context canceled")
}

type fakeSchedulerApp struct {
	queue []*Run
	runs  []*Run
	Application
}

func (f *fakeSchedulerApp) ListRuns(context.Context, RunListOptions) (*RunList, error) {
	return &RunList{Items: f.runs, Pagination: NewPagination(ListOptions{}, len(f.runs))}, nil
}

func (f *fakeSchedulerApp) EnqueuePlan(ctx context.Context, runID string) (*Run, error) {
	for _, run := range f.runs {
		if run.ID() == runID {
			run.status = RunPlanQueued
			return run, nil
		}
	}
	return nil, nil
}

func (f *fakeSchedulerApp) UpdateWorkspaceQueue(run *Run) error {
	for pos, queued := range f.queue {
		if queued.ID() == run.ID() {
			f.queue[pos] = run
			return nil
		}
	}
	f.queue = append(f.queue, run)
	return nil
}

func (f *fakeSchedulerApp) GetWorkspaceQueue(workspaceID string) ([]*Run, error) {
	return f.queue, nil
}

func (f *fakeSchedulerApp) UnlockWorkspace(context.Context, WorkspaceSpec, WorkspaceUnlockOptions) (*Workspace, error) {
	return nil, nil
}

func (f *fakeSchedulerApp) Watch(context.Context, WatchOptions) (<-chan Event, error) {
	return make(chan Event), nil
}
