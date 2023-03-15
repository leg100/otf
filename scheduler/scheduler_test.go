package scheduler

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
)

// TestScheduler checks the scheduler is creating workspace queues and
// forwarding events to the queue handlers.
func TestScheduler(t *testing.T) {
	t.Run("create workspace queue from db", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ws := &workspace.Workspace{ID: "ws-123"}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, nil)
		go scheduler.reinitialize(ctx)

		assert.Equal(t, otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, 1, len(scheduler.queues))

		cancel()
	})

	t.Run("create workspace queue from event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ws := &workspace.Workspace{ID: "ws-123"}
		event := otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws}
		scheduler, got := newTestScheduler(nil, nil, event)
		go scheduler.reinitialize(ctx)
		assert.Equal(t, event, <-got)
		assert.Equal(t, 1, len(scheduler.queues))
		cancel()
	})

	t.Run("delete workspace queue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		// ws is to be created and then deleted
		ws := &workspace.Workspace{ID: "ws-123"}
		del := otf.Event{Type: otf.EventWorkspaceDeleted, Payload: ws}
		// necessary so that we can synchronise test below
		sync := otf.Event{Payload: &workspace.Workspace{ID: "ws-123"}}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, nil, del, sync)
		go scheduler.reinitialize(ctx)

		assert.Equal(t, otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, sync, <-got)
		assert.NotContains(t, scheduler.queues, ws)

		cancel()
	})

	t.Run("relay run from db", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ws := &workspace.Workspace{ID: "ws-123"}
		r := &run.Run{WorkspaceID: "ws-123"}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, []*run.Run{r})
		go scheduler.reinitialize(ctx)

		assert.Equal(t, otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, otf.Event{Type: otf.EventRunStatusUpdate, Payload: r}, <-got)

		cancel()
	})

	t.Run("relay run from event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ws := &workspace.Workspace{ID: "ws-123"}
		event := otf.Event{Payload: &run.Run{WorkspaceID: "ws-123"}}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, nil, event)
		go scheduler.reinitialize(ctx)

		assert.Equal(t, otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, event, <-got)

		cancel()
	})

	t.Run("relay runs in reverse order", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ws := &workspace.Workspace{ID: "ws-123"}
		run1 := &run.Run{WorkspaceID: "ws-123"}
		run2 := &run.Run{WorkspaceID: "ws-123"}
		scheduler, got := newTestScheduler([]*workspace.Workspace{ws}, []*run.Run{run1, run2})
		go scheduler.reinitialize(ctx)

		assert.Equal(t, otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, otf.Event{Type: otf.EventRunStatusUpdate, Payload: run2}, <-got)
		assert.Equal(t, otf.Event{Type: otf.EventRunStatusUpdate, Payload: run1}, <-got)

		cancel()
	})
}

func newTestScheduler(workspaces []*workspace.Workspace, runs []*run.Run, events ...otf.Event) (*scheduler, <-chan otf.Event) {
	ch := make(chan otf.Event, len(events))
	for _, ev := range events {
		ch <- ev
	}

	// construct and run scheduler
	scheduler := Start(logr.Discard(), &fakeSchedulerApp{
		runs:       runs,
		workspaces: workspaces,
		events:     ch,
	})
	// handled chan receives events relayed to handlers
	handled := make(chan otf.Event)
	scheduler.queueFactory = &fakeQueueFactory{events: handled}
	return scheduler, handled
}

type fakeSchedulerApp struct {
	runs       []*run.Run
	workspaces []*workspace.Workspace
	events     chan otf.Event

	otf.Application
}

func (f *fakeSchedulerApp) ListRuns(context.Context, run.RunListOptions) (*run.RunList, error) {
	return &run.RunList{
		Items:      f.runs,
		Pagination: otf.NewPagination(otf.ListOptions{}, len(f.runs)),
	}, nil
}

func (f *fakeSchedulerApp) ListWorkspaces(context.Context, workspace.WorkspaceListOptions) (*workspace.WorkspaceList, error) {
	return &workspace.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(otf.ListOptions{}, len(f.workspaces)),
	}, nil
}

func (f *fakeSchedulerApp) Watch(context.Context, otf.WatchOptions) (<-chan otf.Event, error) {
	return f.events, nil
}

type fakeQueueFactory struct {
	events chan otf.Event
}

func (f *fakeQueueFactory) newQueue(otf.Application, logr.Logger, *workspace.Workspace) eventHandler {
	return &fakeQueue{events: f.events}
}

type fakeQueue struct {
	events chan otf.Event
}

func (f *fakeQueue) handleEvent(ctx context.Context, event otf.Event) error {
	f.events <- event
	return nil
}
