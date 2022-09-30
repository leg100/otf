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
	ctx, cancel := context.WithCancel(context.Background())

	org := newTestOrganization(t)
	ws1 := newTestWorkspace(t, org)
	ws2 := newTestWorkspace(t, org)
	ws3 := newTestWorkspace(t, org)
	cv := newTestConfigurationVersion(t, ws1, ConfigurationVersionCreateOptions{})

	// run[1-2] are in the DB; run[3-4] are events
	run1 := NewRun(cv, ws1, RunCreateOptions{})
	run2 := NewRun(cv, ws2, RunCreateOptions{})
	run3 := NewRun(cv, ws1, RunCreateOptions{})
	run4 := NewRun(cv, ws2, RunCreateOptions{})

	db := []*Run{run1, run2}
	events := make(chan Event, 3)
	events <- Event{Type: EventRunCreated, Payload: run3}
	events <- Event{Type: EventRunCreated, Payload: run4}
	events <- Event{Type: EventWorkspaceCreated, Payload: ws3}

	// handled chan receives events relaayed to handlers
	handled := make(chan Event)
	// construct and run scheduler
	app := &fakeSchedulerApp{
		db:     db,
		events: events,
	}
	scheduler := NewScheduler(logr.Discard(), app)
	scheduler.workspaceQueueFactory = &fakeQueueMaker{events: handled}
	go scheduler.reinitialize(ctx)

	// scheduler reverses order of runs retrieved from DB
	got2 := <-handled
	assert.Equal(t, Event{Type: EventRunStatusUpdate, Payload: run2}, got2)

	got1 := <-handled
	assert.Equal(t, Event{Type: EventRunStatusUpdate, Payload: run1}, got1)

	got3 := <-handled
	assert.Equal(t, Event{Type: EventRunCreated, Payload: run3}, got3)

	got4 := <-handled
	assert.Equal(t, Event{Type: EventRunCreated, Payload: run4}, got4)

	got5 := <-handled
	assert.Equal(t, Event{Type: EventWorkspaceCreated, Payload: ws3}, got5)

	// should be three queues, one for each workspace
	assert.Equal(t, 3, len(scheduler.queues))

	cancel()
}

type fakeSchedulerApp struct {
	db     []*Run
	events chan Event

	Application
}

func (f *fakeSchedulerApp) GetWorkspace(ctx context.Context, spec WorkspaceSpec) (*Workspace, error) {
	return &Workspace{id: *spec.ID}, nil
}

func (f *fakeSchedulerApp) ListRuns(context.Context, RunListOptions) (*RunList, error) {
	return &RunList{
		Items:      f.db,
		Pagination: NewPagination(ListOptions{}, len(f.db)),
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
