package scheduler

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	org := otf.NewTestOrganization(t)

	t.Run("handle several runs", func(t *testing.T) {
		ws := otf.NewTestWorkspace(t, org)
		cv1 := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})
		run1 := otf.NewRun(cv1, ws, otf.RunCreateOptions{})
		run2 := otf.NewRun(cv1, ws, otf.RunCreateOptions{})
		run3 := otf.NewRun(cv1, ws, otf.RunCreateOptions{})
		app := newFakeQueueApp(ws, run1, run2, run3)
		q := newTestQueue(app, ws)

		// enqueue run1, check it is current run
		err := q.handleEvent(ctx, otf.Event{Payload: run1})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run1.ID(), q.current.ID())
		assert.True(t, q.ws.Locked())

		// enqueue run2, check it is in queue
		err = q.handleEvent(ctx, otf.Event{Payload: run2})
		require.NoError(t, err)
		assert.Equal(t, 1, len(q.queue))
		assert.Equal(t, run2.ID(), q.queue[0].ID())
		assert.True(t, q.ws.Locked())

		// enqueue run3, check it is in queue
		err = q.handleEvent(ctx, otf.Event{Payload: run3})
		require.NoError(t, err)
		assert.Equal(t, 2, len(q.queue))
		assert.Equal(t, run3.ID(), q.queue[1].ID())
		assert.True(t, q.ws.Locked())

		// cancel run2, check it is removed from queue and run3 is shuffled forward
		_, err = run2.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, otf.Event{Payload: run2})
		require.NoError(t, err)
		assert.Equal(t, 1, len(q.queue))
		assert.Equal(t, run3.ID(), q.queue[0].ID())
		assert.True(t, q.ws.Locked())

		// cancel run1; check run3 takes its place as current run
		_, err = run1.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, otf.Event{Payload: run1})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run3.ID(), q.current.ID())
		assert.True(t, q.ws.Locked())

		// cancel run3; check everything is empty and workspace is unlocked
		_, err = run3.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, otf.Event{Payload: run3})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Nil(t, q.current)
		assert.False(t, q.ws.Locked())
	})

	t.Run("speculative run", func(t *testing.T) {
		ws := otf.NewTestWorkspace(t, org)
		cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{
			Speculative: otf.Bool(true),
		})
		run := otf.NewRun(cv, ws, otf.RunCreateOptions{})
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleEvent(ctx, otf.Event{Payload: run})
		require.NoError(t, err)
		// should be scheduled but not enqueued onto workspace q
		assert.Contains(t, app.scheduled, run.ID())
		assert.Equal(t, 0, len(q.queue))
	})

	t.Run("user locked", func(t *testing.T) {
		ws := otf.NewTestWorkspace(t, org)
		cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})
		run := otf.NewRun(cv, ws, otf.RunCreateOptions{})
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		// user locks workspace; new run should be made the current run but should not
		// be scheduled nor replace the workspace lock
		bob := otf.NewUser("bob")
		err := ws.Lock(bob)
		require.NoError(t, err)
		err = q.handleEvent(ctx, otf.Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run.ID(), q.current.ID())
		assert.Equal(t, bob, q.ws.GetLock())

		// user unlocks workspace; run should be scheduled, locking the workspace
		err = ws.Unlock(bob, false)
		require.NoError(t, err)
		err = q.handleEvent(ctx, otf.Event{Type: otf.EventWorkspaceUnlocked, Payload: ws})
		require.NoError(t, err)
		assert.Equal(t, run.ID(), q.current.ID())
		assert.Equal(t, run.ID(), q.ws.GetLock().ID())
	})

	t.Run("do not schedule non-pending run", func(t *testing.T) {
		ws := otf.NewTestWorkspace(t, org)
		run := otf.NewTestRun(t, otf.TestRunCreateOptions{
			Status:    otf.RunPlanning,
			Workspace: ws,
		})
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleEvent(ctx, otf.Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, run.ID(), q.current.ID())
		assert.NotContains(t, q.Application.(*fakeQueueApp).scheduled, run.ID())
	})

	t.Run("do not set current run if already latest run on workspace", func(t *testing.T) {
		ws := otf.NewTestWorkspace(t, org)
		cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})
		run := otf.NewRun(cv, ws, otf.RunCreateOptions{})
		ws.SetLatestRun(run.ID())
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleEvent(ctx, otf.Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, run.ID(), q.current.ID())
		assert.NotContains(t, app.current, run.ID())
	})
}

func newTestQueue(app *fakeQueueApp, ws *otf.Workspace) *queue {
	return &queue{
		ws:          ws,
		Application: app,
		Logger:      logr.Discard(),
	}
}

type fakeQueueApp struct {
	ws        *otf.Workspace
	runs      map[string]*otf.Run // mock run db
	scheduled []string            // list of IDs of scheduled runs
	current   []string            // list of IDs of runs that have been set as the current run

	otf.Application
}

func newFakeQueueApp(ws *otf.Workspace, runs ...*otf.Run) *fakeQueueApp {
	db := make(map[string]*otf.Run, len(runs))
	for _, r := range runs {
		db[r.ID()] = r
	}
	return &fakeQueueApp{ws: ws, runs: db}
}

func (f *fakeQueueApp) EnqueuePlan(ctx context.Context, runID string) (*otf.Run, error) {
	f.scheduled = append(f.scheduled, runID)
	return f.runs[runID], nil
}

func (f *fakeQueueApp) LockWorkspace(ctx context.Context, workspaceID string, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	subj, err := otf.LockFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := f.ws.Lock(subj); err != nil {
		return nil, err
	}
	return f.ws, nil
}

func (f *fakeQueueApp) UnlockWorkspace(ctx context.Context, workspaceID string, opts otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	subj, err := otf.LockFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := f.ws.Unlock(subj, false); err != nil {
		return nil, err
	}
	return f.ws, nil
}

func (f *fakeQueueApp) SetCurrentRun(ctx context.Context, workspaceID, runID string) (*otf.Workspace, error) {
	f.ws.SetLatestRun(runID)
	return f.ws, nil
}
