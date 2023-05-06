package scheduler

import (
	"context"
	"testing"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
)

// TestScheduler checks the scheduler is creating workspace queues and
// forwarding events to the queue handlers.
func TestScheduler(t *testing.T) {
	t.Run("create workspace queue from db", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws := &workspace.Workspace{ID: "ws-123"}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, nil)
		go scheduler.reinitialize(ctx)

		assert.Equal(t, internal.Event{Type: internal.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, 1, len(scheduler.queues))
	})

	t.Run("create workspace queue from event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws := &workspace.Workspace{ID: "ws-123"}
		event := internal.Event{Type: internal.EventWorkspaceCreated, Payload: ws}
		scheduler, got := newTestScheduler(nil, nil, event)
		go scheduler.reinitialize(ctx)
		assert.Equal(t, event, <-got)
		assert.Equal(t, 1, len(scheduler.queues))
	})

	t.Run("delete workspace queue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// ws is to be created and then deleted
		ws := &workspace.Workspace{ID: "ws-123"}
		del := internal.Event{Type: internal.EventWorkspaceDeleted, Payload: ws}
		// necessary so that we can synchronise test below
		sync := internal.Event{Payload: &workspace.Workspace{ID: "ws-123"}}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, nil, del, sync)
		go scheduler.reinitialize(ctx)

		assert.Equal(t, internal.Event{Type: internal.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, sync, <-got)
		assert.NotContains(t, scheduler.queues, ws)
	})

	t.Run("relay run from db", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws := &workspace.Workspace{ID: "ws-123"}
		r := &run.Run{WorkspaceID: "ws-123"}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, []*run.Run{r})
		go scheduler.reinitialize(ctx)

		assert.Equal(t, internal.Event{Type: internal.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, internal.Event{Type: internal.EventRunStatusUpdate, Payload: r}, <-got)
	})

	t.Run("relay run from event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws := &workspace.Workspace{ID: "ws-123"}
		event := internal.Event{Payload: &run.Run{WorkspaceID: "ws-123"}}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, nil, event)
		go scheduler.reinitialize(ctx)

		assert.Equal(t, internal.Event{Type: internal.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, event, <-got)
	})

	t.Run("relay runs in reverse order", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws := &workspace.Workspace{ID: "ws-123"}
		run1 := &run.Run{WorkspaceID: "ws-123"}
		run2 := &run.Run{WorkspaceID: "ws-123"}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, []*run.Run{run1, run2})
		go scheduler.reinitialize(ctx)

		assert.Equal(t, internal.Event{Type: internal.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, internal.Event{Type: internal.EventRunStatusUpdate, Payload: run2}, <-got)
		assert.Equal(t, internal.Event{Type: internal.EventRunStatusUpdate, Payload: run1}, <-got)
	})
}
