package scheduler

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("handle several runs", func(t *testing.T) {
		ws := &workspace.Workspace{ID: "ws-123"}
		run1 := &run.Run{ID: "run-1", WorkspaceID: "ws-123", Status: otf.RunPending}
		run2 := &run.Run{ID: "run-2", WorkspaceID: "ws-123", Status: otf.RunPending}
		run3 := &run.Run{ID: "run-3", WorkspaceID: "ws-123", Status: otf.RunPending}
		app := newFakeQueueApp(ws, run1, run2, run3)
		q := newTestQueue(app, ws)

		// enqueue run1, check it is current run
		err := q.handleEvent(ctx, otf.Event{Payload: run1})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run1.ID, q.current.ID)
		assert.True(t, q.ws.Locked())

		// enqueue run2, check it is in queue
		err = q.handleEvent(ctx, otf.Event{Payload: run2})
		require.NoError(t, err)
		if assert.Equal(t, 1, len(q.queue)) {
			assert.Equal(t, run2.ID, q.queue[0].ID)
		}
		assert.True(t, q.ws.Locked())

		// enqueue run3, check it is in queue
		err = q.handleEvent(ctx, otf.Event{Payload: run3})
		require.NoError(t, err)
		if assert.Equal(t, 2, len(q.queue)) {
			assert.Equal(t, run3.ID, q.queue[1].ID)
		}
		assert.True(t, q.ws.Locked())

		// cancel run2, check it is removed from queue and run3 is shuffled forward
		_, err = run2.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, otf.Event{Payload: run2})
		require.NoError(t, err)
		if assert.Equal(t, 1, len(q.queue)) {
			assert.Equal(t, run3.ID, q.queue[0].ID)
		}
		assert.True(t, q.ws.Locked())

		// cancel run1; check run3 takes its place as current run
		_, err = run1.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, otf.Event{Payload: run1})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run3.ID, q.current.ID)
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
		ws := &workspace.Workspace{ID: "ws-123"}
		run := &run.Run{Status: otf.RunPending, WorkspaceID: "ws-123", Speculative: true}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleEvent(ctx, otf.Event{Payload: run})
		require.NoError(t, err)
		// should be scheduled but not enqueued onto workspace q
		assert.Equal(t, otf.RunPlanQueued, run.Status)
		assert.Equal(t, 0, len(q.queue))
	})

	t.Run("user locked", func(t *testing.T) {
		ws := &workspace.Workspace{ID: "ws-123"}
		run := &run.Run{ID: "run-123", WorkspaceID: "ws-123", Status: otf.RunPending}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		// user locks workspace; new run should be made the current run but should not
		// be scheduled nor replace the workspace lock
		lock := workspace.UserLock{}
		err := ws.Lock.Lock(lock)
		require.NoError(t, err)
		err = q.handleEvent(ctx, otf.Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run.ID, q.current.ID)
		assert.Equal(t, lock, q.ws.Lock.LockedState)

		// user unlocks workspace; run should be scheduled, locking the workspace
		err = ws.Unlock(lock, false)
		require.NoError(t, err)
		err = q.handleEvent(ctx, otf.Event{Type: workspace.EventUnlocked, Payload: ws})
		require.NoError(t, err)
		assert.Equal(t, run.ID, q.current.ID)
		assert.Equal(t, workspace.RunLock{ID: "run-123"}, q.ws.Lock.LockedState)
	})

	t.Run("do not schedule non-pending run", func(t *testing.T) {
		ws := &workspace.Workspace{ID: "ws-123"}
		run := &run.Run{WorkspaceID: "ws-123", Status: otf.RunPlanning}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleEvent(ctx, otf.Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, run.ID, q.current.ID)
		assert.Equal(t, otf.RunPlanning, run.Status)
	})

	t.Run("do not set current run if already latest run on workspace", func(t *testing.T) {
		run := &run.Run{WorkspaceID: "ws-123"}
		ws := &workspace.Workspace{ID: "ws-123", LatestRunID: &run.ID}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleEvent(ctx, otf.Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, run.ID, q.current.ID)
		assert.NotContains(t, app.current, run.ID)
	})
}

func newTestQueue(services *fakeQueueServices, ws *workspace.Workspace) *queue {
	return &queue{
		WorkspaceService: services,
		RunService:       services,
		ws:               ws,
		Logger:           logr.Discard(),
	}
}

type fakeQueueServices struct {
	ws        *workspace.Workspace
	runs      map[string]*run.Run // mock run db
	scheduled []string            // list of IDs of scheduled runs
	current   []string            // list of IDs of runs that have been set as the current run

	WorkspaceService
	RunService
}

func newFakeQueueApp(ws *workspace.Workspace, runs ...*run.Run) *fakeQueueServices {
	db := make(map[string]*run.Run, len(runs))
	for _, r := range runs {
		db[r.ID] = r
	}
	return &fakeQueueServices{ws: ws, runs: db}
}

func (f *fakeQueueServices) EnqueuePlan(ctx context.Context, runID string) (*run.Run, error) {
	f.runs[runID].Status = otf.RunPlanQueued
	return f.runs[runID], nil
}

func (f *fakeQueueServices) LockWorkspace(ctx context.Context, workspaceID string, runID *string) (*workspace.Workspace, error) {
	state, err := workspace.GetLockedState(nil, runID)
	if err != nil {
		return nil, err
	}
	if err := f.ws.Lock.Lock(state); err != nil {
		return nil, err
	}
	return f.ws, nil
}

func (f *fakeQueueServices) UnlockWorkspace(ctx context.Context, workspaceID string, runID *string, force bool) (*workspace.Workspace, error) {
	state, err := workspace.GetLockedState(nil, runID)
	if err != nil {
		return nil, err
	}
	if err := f.ws.Lock.Unlock(state, false); err != nil {
		return nil, err
	}
	return f.ws, nil
}

func (f *fakeQueueServices) SetCurrentRun(ctx context.Context, workspaceID, runID string) (*workspace.Workspace, error) {
	f.ws.LatestRunID = &runID
	return f.ws, nil
}
