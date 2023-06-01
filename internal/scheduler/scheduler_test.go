package scheduler

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
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
		go scheduler.Start(ctx)

		assert.Equal(t, pubsub.Event{Type: pubsub.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, 1, len(scheduler.queues))
	})

	t.Run("create workspace queue from event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws := &workspace.Workspace{ID: "ws-123"}
		event := pubsub.Event{Type: pubsub.EventWorkspaceCreated, Payload: ws}
		scheduler, got := newTestScheduler(nil, nil, event)
		go scheduler.Start(ctx)
		assert.Equal(t, event, <-got)
		assert.Equal(t, 1, len(scheduler.queues))
	})

	t.Run("delete workspace queue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// ws is to be created and then deleted
		ws := &workspace.Workspace{ID: "ws-123"}
		del := pubsub.Event{Type: pubsub.DeletedEvent, Payload: ws}
		// necessary so that we can synchronise test below
		sync := pubsub.Event{Payload: &workspace.Workspace{ID: "ws-123"}}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, nil, del, sync)
		go scheduler.Start(ctx)

		assert.Equal(t, pubsub.Event{Type: pubsub.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, sync, <-got)
		assert.NotContains(t, scheduler.queues, ws)
	})

	t.Run("relay run from db", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws := &workspace.Workspace{ID: "ws-123"}
		r := &run.Run{WorkspaceID: "ws-123"}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, []*run.Run{r})
		go scheduler.Start(ctx)

		assert.Equal(t, pubsub.Event{Type: pubsub.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, pubsub.Event{Type: pubsub.EventRunStatusUpdate, Payload: r}, <-got)
	})

	t.Run("relay run from event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws := &workspace.Workspace{ID: "ws-123"}
		event := pubsub.Event{Payload: &run.Run{WorkspaceID: "ws-123"}}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, nil, event)
		go scheduler.Start(ctx)

		assert.Equal(t, pubsub.Event{Type: pubsub.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, event, <-got)
	})

	t.Run("relay runs in reverse order", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ws := &workspace.Workspace{ID: "ws-123"}
		run1 := &run.Run{WorkspaceID: "ws-123"}
		run2 := &run.Run{WorkspaceID: "ws-123"}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, []*run.Run{run1, run2})
		go scheduler.Start(ctx)

		assert.Equal(t, pubsub.Event{Type: pubsub.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, pubsub.Event{Type: pubsub.EventRunStatusUpdate, Payload: run2}, <-got)
		assert.Equal(t, pubsub.Event{Type: pubsub.EventRunStatusUpdate, Payload: run1}, <-got)
	})
}
