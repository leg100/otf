package otf

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

// TestScheduler checks the scheduler is creating workspace queues and
// forwarding events to the queue handlers.
func TestScheduler(t *testing.T) {
	org := NewTestOrganization(t)

	t.Run("create workspace queue from db", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ws := NewTestWorkspace(t, org, WorkspaceCreateOptions{})
		scheduler, got := newTestScheduler([]*Workspace{ws}, nil)
		go scheduler.reinitialize(ctx)

		assert.Equal(t, Event{Type: EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, 1, len(scheduler.queues))

		cancel()
	})

	t.Run("create workspace queue from event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		event := Event{Type: EventWorkspaceCreated, Payload: NewTestWorkspace(t, org, WorkspaceCreateOptions{})}
		scheduler, got := newTestScheduler(nil, nil, event)
		go scheduler.reinitialize(ctx)
		assert.Equal(t, event, <-got)
		assert.Equal(t, 1, len(scheduler.queues))
		cancel()
	})

	t.Run("delete workspace queue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		// ws is to be created and then deleted
		ws := NewTestWorkspace(t, org, WorkspaceCreateOptions{})
		del := Event{Type: EventWorkspaceDeleted, Payload: ws}
		// necessary so that we can synchronise test below
		sync := Event{Payload: NewTestWorkspace(t, org, WorkspaceCreateOptions{})}
		scheduler, got := newTestScheduler([]*Workspace{ws}, nil, del, sync)
		go scheduler.reinitialize(ctx)

		assert.Equal(t, Event{Type: EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, sync, <-got)
		assert.NotContains(t, scheduler.queues, ws)

		cancel()
	})

	t.Run("relay run from db", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ws := NewTestWorkspace(t, org, WorkspaceCreateOptions{})
		cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{})
		run := NewRun(cv, ws, RunCreateOptions{})
		scheduler, got := newTestScheduler([]*Workspace{ws}, []*Run{run})
		go scheduler.reinitialize(ctx)

		assert.Equal(t, Event{Type: EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, Event{Type: EventRunStatusUpdate, Payload: run}, <-got)

		cancel()
	})

	t.Run("relay run from event", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ws := NewTestWorkspace(t, org, WorkspaceCreateOptions{})
		cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{})
		event := Event{Payload: NewRun(cv, ws, RunCreateOptions{})}
		scheduler, got := newTestScheduler([]*Workspace{ws}, nil, event)
		go scheduler.reinitialize(ctx)

		assert.Equal(t, Event{Type: EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, event, <-got)

		cancel()
	})

	t.Run("relay runs in reverse order", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ws := NewTestWorkspace(t, org, WorkspaceCreateOptions{})
		cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{})
		run1 := NewRun(cv, ws, RunCreateOptions{})
		run2 := NewRun(cv, ws, RunCreateOptions{})
		scheduler, got := newTestScheduler([]*Workspace{ws}, []*Run{run1, run2})
		go scheduler.reinitialize(ctx)

		assert.Equal(t, Event{Type: EventWorkspaceCreated, Payload: ws}, <-got)
		assert.Equal(t, Event{Type: EventRunStatusUpdate, Payload: run2}, <-got)
		assert.Equal(t, Event{Type: EventRunStatusUpdate, Payload: run1}, <-got)

		cancel()
	})
}

func newTestScheduler(workspaces []*Workspace, runs []*Run, events ...Event) (*Scheduler, <-chan Event) {
	ch := make(chan Event, len(events))
	for _, ev := range events {
		ch <- ev
	}

	// construct and run scheduler
	scheduler := NewScheduler(logr.Discard(), &fakeSchedulerApp{
		runs:       runs,
		workspaces: workspaces,
		events:     ch,
	})
	// handled chan receives events relayed to handlers
	handled := make(chan Event)
	scheduler.workspaceQueueFactory = &fakeQueueMaker{events: handled}
	return scheduler, handled
}

type fakeSchedulerApp struct {
	runs       []*Run
	workspaces []*Workspace
	events     chan Event

	Application
}

func (f *fakeSchedulerApp) ListRuns(context.Context, RunListOptions) (*RunList, error) {
	return &RunList{
		Items:      f.runs,
		Pagination: NewPagination(ListOptions{}, len(f.runs)),
	}, nil
}

func (f *fakeSchedulerApp) ListWorkspaces(context.Context, WorkspaceListOptions) (*WorkspaceList, error) {
	return &WorkspaceList{
		Items:      f.workspaces,
		Pagination: NewPagination(ListOptions{}, len(f.workspaces)),
	}, nil
}

func (f *fakeSchedulerApp) Watch(context.Context, WatchOptions) (<-chan Event, error) {
	return f.events, nil
}

type fakeQueueMaker struct {
	events chan Event
}

func (f *fakeQueueMaker) NewWorkspaceQueue(Application, logr.Logger, *Workspace) eventHandler {
	return &fakeWorkspaceQueue{events: f.events}
}

type fakeWorkspaceQueue struct {
	events chan Event
}

func (f *fakeWorkspaceQueue) handleEvent(ctx context.Context, event Event) error {
	f.events <- event
	return nil
}
