package scheduler

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/resource"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wsID := resource.NewID(resource.WorkspaceKind)
	userID := resource.NewID(resource.UserKind)

	t.Run("handle several runs", func(t *testing.T) {
		ws := &workspace.Workspace{ID: wsID}
		run1 := &otfrun.Run{ID: resource.NewID(resource.RunKind), WorkspaceID: wsID, Status: otfrun.RunPending}
		run2 := &otfrun.Run{ID: resource.NewID(resource.RunKind), WorkspaceID: wsID, Status: otfrun.RunPending}
		run3 := &otfrun.Run{ID: resource.NewID(resource.RunKind), WorkspaceID: wsID, Status: otfrun.RunPending}
		app := newFakeQueueApp(ws, run1, run2, run3)
		q := newTestQueue(app, ws)

		// enqueue run1, check it is current run
		err := q.handleRun(ctx, run1)
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run1.ID, q.current.ID)
		assert.True(t, q.ws.Lock.Locked())

		// enqueue run2, check it is in queue
		err = q.handleRun(ctx, run2)
		require.NoError(t, err)
		if assert.Equal(t, 1, len(q.queue)) {
			assert.Equal(t, run2.ID, q.queue[0].ID)
		}
		assert.True(t, q.ws.Lock.Locked())

		// enqueue run3, check it is in queue
		err = q.handleRun(ctx, run3)
		require.NoError(t, err)
		if assert.Equal(t, 2, len(q.queue)) {
			assert.Equal(t, run3.ID, q.queue[1].ID)
		}
		assert.True(t, q.ws.Lock.Locked())

		// cancel run2, check it is removed from queue and run3 is shuffled forward
		err = run2.Cancel(false, false)
		require.NoError(t, err)
		err = q.handleRun(ctx, run2)
		require.NoError(t, err)
		if assert.Equal(t, 1, len(q.queue)) {
			assert.Equal(t, run3.ID, q.queue[0].ID)
		}
		assert.True(t, q.ws.Lock.Locked())

		// cancel run1; check run3 takes its place as current run
		err = run1.Cancel(false, false)
		require.NoError(t, err)
		err = q.handleRun(ctx, run1)
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run3.ID, q.current.ID)
		assert.True(t, q.ws.Lock.Locked())

		// cancel run3; check everything is empty and workspace is unlocked
		err = run3.Cancel(false, false)
		require.NoError(t, err)
		err = q.handleRun(ctx, run3)
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Nil(t, q.current)
		assert.False(t, q.ws.Lock.Locked())
	})

	t.Run("speculative run", func(t *testing.T) {
		ws := &workspace.Workspace{ID: wsID}
		run := &otfrun.Run{Status: otfrun.RunPending, WorkspaceID: wsID, PlanOnly: true}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleRun(ctx, run)
		require.NoError(t, err)
		// should be scheduled but not enqueued onto workspace q
		assert.Equal(t, otfrun.RunPlanQueued, run.Status)
		assert.Equal(t, 0, len(q.queue))
	})

	t.Run("user locked", func(t *testing.T) {
		ws := &workspace.Workspace{ID: wsID}
		run := &otfrun.Run{ID: resource.NewID(resource.RunKind), WorkspaceID: wsID, Status: otfrun.RunPending}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		// user locks workspace; new run should be made the current run but should not
		// be scheduled nor replace the user lock
		err := ws.Lock.Enlock(userID)
		require.NoError(t, err)
		err = q.handleRun(ctx, run)
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run.ID, q.current.ID)
		assert.Equal(t, resource.UserKind, q.ws.Lock.Kind)

		// user unlocks workspace; run should be scheduled, locking the workspace
		err = ws.Lock.Unlock(userID, false)
		require.NoError(t, err)
		err = q.handleWorkspace(ctx, ws)
		require.NoError(t, err)
		assert.Equal(t, run.ID, q.current.ID)
		assert.Equal(t, resource.RunKind, q.ws.Lock.Kind)
	})

	t.Run("do not schedule non-pending run", func(t *testing.T) {
		ws := &workspace.Workspace{ID: wsID}
		run := &otfrun.Run{WorkspaceID: wsID, Status: otfrun.RunPlanning}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleRun(ctx, run)
		require.NoError(t, err)
		assert.Equal(t, run.ID, q.current.ID)
		assert.Equal(t, otfrun.RunPlanning, run.Status)
	})

	t.Run("do not set current run if already latest run on workspace", func(t *testing.T) {
		run := &otfrun.Run{WorkspaceID: wsID}
		ws := &workspace.Workspace{ID: wsID, LatestRun: &workspace.LatestRun{ID: run.ID}}
		app := newFakeQueueApp(ws, run)
		q := newTestQueue(app, ws)

		err := q.handleRun(ctx, run)
		require.NoError(t, err)
		assert.Equal(t, run.ID, q.current.ID)
		assert.NotContains(t, app.current, run.ID)
	})
}

func newTestQueue(services *fakeQueueServices, ws *workspace.Workspace) *queue {
	return &queue{
		workspaceClient: &fakeWorkspaceService{
			ws: ws,
		},
		runClient: services,
		ws:        ws,
		Logger:    logr.Discard(),
	}
}

type fakeQueueServices struct {
	ws      *workspace.Workspace
	runs    map[resource.ID]*otfrun.Run // mock run db
	current []resource.ID               // list of IDs of runs that have been set as the current run

	runClient
}

func newFakeQueueApp(ws *workspace.Workspace, runs ...*otfrun.Run) *fakeQueueServices {
	db := make(map[resource.ID]*otfrun.Run, len(runs))
	for _, r := range runs {
		db[r.ID] = r
	}
	return &fakeQueueServices{ws: ws, runs: db}
}

func (f *fakeQueueServices) EnqueuePlan(ctx context.Context, runID resource.ID) (*otfrun.Run, error) {
	f.runs[runID].Status = otfrun.RunPlanQueued
	return f.runs[runID], nil
}

type fakeWorkspaceService struct {
	ws *workspace.Workspace

	// fakeWorkspaceService does not implement all of workspaceClient
	workspaceClient
}

func (f *fakeWorkspaceService) Lock(ctx context.Context, workspaceID resource.ID, runID *resource.ID) (*workspace.Workspace, error) {
	if err := f.ws.Lock.Enlock(*runID); err != nil {
		return nil, err
	}
	return f.ws, nil
}

func (f *fakeWorkspaceService) Unlock(ctx context.Context, workspaceID resource.ID, runID *resource.ID, force bool) (*workspace.Workspace, error) {
	if err := f.ws.Lock.Unlock(*runID, false); err != nil {
		return nil, err
	}
	return f.ws, nil
}

func (f *fakeWorkspaceService) SetCurrentRun(ctx context.Context, workspaceID, runID resource.ID) (*workspace.Workspace, error) {
	f.ws.LatestRun = &workspace.LatestRun{ID: runID}
	return f.ws, nil
}
