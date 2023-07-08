package scheduler

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("handle several runs", func(t *testing.T) {
		ws := &workspace.Workspace{ID: "ws-123"}
		run1 := &run.Run{ID: "run-1", WorkspaceID: "ws-123", Status: internal.RunPending}
		run2 := &run.Run{ID: "run-2", WorkspaceID: "ws-123", Status: internal.RunPending}
		run3 := &run.Run{ID: "run-3", WorkspaceID: "ws-123", Status: internal.RunPending}
		app := newFakeQueueApp(ws, run1, run2, run3)
		q := newTestQueue(app, ws)

		// enqueue run1, check it is current run
		err := q.handleEvent(ctx, pubsub.Event{Payload: run1})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run1.ID, q.current.ID)
		assert.True(t, q.ws.Locked())

		// enqueue run2, check it is in queue
		err = q.handleEvent(ctx, pubsub.Event{Payload: run2})
		require.NoError(t, err)
		if assert.Equal(t, 1, len(q.queue)) {
			assert.Equal(t, run2.ID, q.queue[0].ID)
		}
		assert.True(t, q.ws.Locked())

		// enqueue run3, check it is in queue
		err = q.handleEvent(ctx, pubsub.Event{Payload: run3})
		require.NoError(t, err)
		if assert.Equal(t, 2, len(q.queue)) {
			assert.Equal(t, run3.ID, q.queue[1].ID)
		}
		assert.True(t, q.ws.Locked())

		// cancel run2, check it is removed from queue and run3 is shuffled forward
		err = run2.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, pubsub.Event{Payload: run2})
		require.NoError(t, err)
		if assert.Equal(t, 1, len(q.queue)) {
			assert.Equal(t, run3.ID, q.queue[0].ID)
		}
		assert.True(t, q.ws.Locked())

		// cancel run1; check run3 takes its place as current run
		err = run1.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, pubsub.Event{Payload: run1})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run3.ID, q.current.ID)
		assert.True(t, q.ws.Locked())

		// cancel run3; check everything is empty and workspace is unlocked
		err = run3.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, pubsub.Event{Payload: run3})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Nil(t, q.current)
		assert.False(t, q.ws.Locked())
	})

	t.Run("speculative run", func(t *testing.T) {
		ws := &workspace.Workspace{ID: "ws-123"}
		run := &run.Run{Status: internal.RunPending, WorkspaceID: "ws-123", PlanOnly: true}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleEvent(ctx, pubsub.Event{Payload: run})
		require.NoError(t, err)
		// should be scheduled but not enqueued onto workspace q
		assert.Equal(t, internal.RunPlanQueued, run.Status)
		assert.Equal(t, 0, len(q.queue))
	})

	t.Run("user locked", func(t *testing.T) {
		ws := &workspace.Workspace{ID: "ws-123"}
		run := &run.Run{ID: "run-123", WorkspaceID: "ws-123", Status: internal.RunPending}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		// user locks workspace; new run should be made the current run but should not
		// be scheduled nor replace the user lock
		err := ws.Enlock("bobby", workspace.UserLock)
		require.NoError(t, err)
		err = q.handleEvent(ctx, pubsub.Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run.ID, q.current.ID)
		assert.Equal(t, workspace.UserLock, q.ws.Lock.LockKind)

		// user unlocks workspace; run should be scheduled, locking the workspace
		err = ws.Unlock("bobby", workspace.UserLock, false)
		require.NoError(t, err)
		err = q.handleEvent(ctx, pubsub.Event{Payload: ws})
		require.NoError(t, err)
		assert.Equal(t, run.ID, q.current.ID)
		assert.Equal(t, workspace.RunLock, q.ws.Lock.LockKind)
	})

	t.Run("do not schedule non-pending run", func(t *testing.T) {
		ws := &workspace.Workspace{ID: "ws-123"}
		run := &run.Run{WorkspaceID: "ws-123", Status: internal.RunPlanning}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleEvent(ctx, pubsub.Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, run.ID, q.current.ID)
		assert.Equal(t, internal.RunPlanning, run.Status)
	})

	t.Run("do not set current run if already latest run on workspace", func(t *testing.T) {
		run := &run.Run{WorkspaceID: "ws-123"}
		ws := &workspace.Workspace{ID: "ws-123", LatestRun: &workspace.LatestRun{ID: run.ID}}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleEvent(ctx, pubsub.Event{Payload: run})
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
	ws      *workspace.Workspace
	runs    map[string]*run.Run // mock run db
	current []string            // list of IDs of runs that have been set as the current run

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
	f.runs[runID].Status = internal.RunPlanQueued
	return f.runs[runID], nil
}

func (f *fakeQueueServices) LockWorkspace(ctx context.Context, workspaceID string, runID *string) (*workspace.Workspace, error) {
	if err := f.ws.Enlock(*runID, workspace.RunLock); err != nil {
		return nil, err
	}
	return f.ws, nil
}

func (f *fakeQueueServices) UnlockWorkspace(ctx context.Context, workspaceID string, runID *string, force bool) (*workspace.Workspace, error) {
	if err := f.ws.Unlock(*runID, workspace.RunLock, false); err != nil {
		return nil, err
	}
	return f.ws, nil
}

func (f *fakeQueueServices) SetCurrentRun(ctx context.Context, workspaceID, runID string) (*workspace.Workspace, error) {
	f.ws.LatestRun = &workspace.LatestRun{ID: runID}
	return f.ws, nil
}
