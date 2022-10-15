package otf

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	org := NewTestOrganization(t)

	t.Run("handle several runs", func(t *testing.T) {
		ws := NewTestWorkspace(t, org)
		q := newTestWorkspaceQueue(ws)
		cv1 := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{})
		run1 := NewRun(cv1, ws, RunCreateOptions{})
		run2 := NewRun(cv1, ws, RunCreateOptions{})
		run3 := NewRun(cv1, ws, RunCreateOptions{})

		// enqueue run1, check it is current run
		err := q.handleEvent(ctx, Event{Payload: run1})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run1.ID(), q.current.ID())
		assert.True(t, q.ws.Locked())

		// enqueue run2, check it is in queue
		err = q.handleEvent(ctx, Event{Payload: run2})
		require.NoError(t, err)
		assert.Equal(t, 1, len(q.queue))
		assert.Equal(t, run2.ID(), q.queue[0].ID())
		assert.True(t, q.ws.Locked())

		// enqueue run3, check it is in queue
		err = q.handleEvent(ctx, Event{Payload: run3})
		require.NoError(t, err)
		assert.Equal(t, 2, len(q.queue))
		assert.Equal(t, run3.ID(), q.queue[1].ID())
		assert.True(t, q.ws.Locked())

		// cancel run2, check it is removed from queue and run3 is shuffled forward
		_, err = run2.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, Event{Payload: run2})
		require.NoError(t, err)
		assert.Equal(t, 1, len(q.queue))
		assert.Equal(t, run3.ID(), q.queue[0].ID())
		assert.True(t, q.ws.Locked())

		// cancel run1; check run3 takes its place as current run
		_, err = run1.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, Event{Payload: run1})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run3.ID(), q.current.ID())
		assert.True(t, q.ws.Locked())

		// cancel run3; check everything is empty and workspace is unlocked
		_, err = run3.Cancel()
		require.NoError(t, err)
		err = q.handleEvent(ctx, Event{Payload: run3})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Nil(t, q.current)
		assert.False(t, q.ws.Locked())
	})

	t.Run("speculative run", func(t *testing.T) {
		ws := NewTestWorkspace(t, org)
		q := newTestWorkspaceQueue(ws)
		cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{
			Speculative: Bool(true),
		})
		run := NewRun(cv, ws, RunCreateOptions{})

		err := q.handleEvent(ctx, Event{Payload: run})
		require.NoError(t, err)
		// should be scheduled but not enqueued onto workspace q
		assert.Contains(t, q.Application.(*fakeWorkspaceQueueApp).scheduled, run.ID())
		assert.Equal(t, 0, len(q.queue))
	})

	t.Run("user locked", func(t *testing.T) {
		ws := NewTestWorkspace(t, org)
		q := newTestWorkspaceQueue(ws)
		cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{})
		run := NewRun(cv, ws, RunCreateOptions{})

		// user locks workspace; new run should be made the current run but should not
		// be scheduled nor replace the workspace lock
		bob := NewUser("bob")
		err := ws.Lock(bob)
		require.NoError(t, err)
		err = q.handleEvent(ctx, Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, 0, len(q.queue))
		assert.Equal(t, run.ID(), q.current.ID())
		assert.Equal(t, bob, q.ws.GetLock())

		// user unlocks workspace; run should be scheduled, locking the workspace
		err = ws.Unlock(bob, false)
		require.NoError(t, err)
		err = q.handleEvent(ctx, Event{Type: EventWorkspaceUnlocked, Payload: ws})
		require.NoError(t, err)
		assert.Equal(t, run.ID(), q.current.ID())
		assert.Equal(t, run.ID(), q.ws.GetLock().ID())
	})

	t.Run("do not schedule non-pending run", func(t *testing.T) {
		ws := NewTestWorkspace(t, org)
		q := newTestWorkspaceQueue(ws)
		cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{})
		run := NewRun(cv, ws, RunCreateOptions{})
		run.status = RunPlanning

		err := q.handleEvent(ctx, Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, run.ID(), q.current.ID())
		assert.NotContains(t, q.Application.(*fakeWorkspaceQueueApp).scheduled, run.ID())
	})

	t.Run("do not set current run if already latest run on workspace", func(t *testing.T) {
		ws := NewTestWorkspace(t, org)
		q := newTestWorkspaceQueue(ws)
		cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{})
		run := NewRun(cv, ws, RunCreateOptions{})
		ws.latestRunID = &run.id

		err := q.handleEvent(ctx, Event{Payload: run})
		require.NoError(t, err)
		assert.Equal(t, run.ID(), q.current.ID())
		assert.NotContains(t, q.Application.(*fakeWorkspaceQueueApp).current, run.ID())
	})
}

func newTestWorkspaceQueue(ws *Workspace) *WorkspaceQueue {
	return &WorkspaceQueue{
		ws:          ws,
		Application: &fakeWorkspaceQueueApp{ws: ws},
		Logger:      logr.Discard(),
	}
}

type fakeWorkspaceQueueApp struct {
	ws *Workspace
	// list of IDs of scheduled runs
	scheduled []string
	// list of IDs of runs that have been set as the current run
	current []string
	Application
}

func (f *fakeWorkspaceQueueApp) EnqueuePlan(ctx context.Context, runID string) (*Run, error) {
	f.scheduled = append(f.scheduled, runID)
	f.ws.lock = &Run{id: runID}
	return &Run{id: runID}, nil
}

func (f *fakeWorkspaceQueueApp) UnlockWorkspace(context.Context, WorkspaceSpec, WorkspaceUnlockOptions) (*Workspace, error) {
	f.ws.lock = &Unlocked{}
	return f.ws, nil
}

func (f *fakeWorkspaceQueueApp) SetCurrentRun(ctx context.Context, workspaceID, runID string) error {
	return nil
}
