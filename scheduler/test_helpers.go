package scheduler

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/workspace"
)

func newTestScheduler(workspaces []*workspace.Workspace, runs []*run.Run, events ...otf.Event) (*scheduler, <-chan otf.Event) {
	ch := make(chan otf.Event, len(events))
	for _, ev := range events {
		ch <- ev
	}

	services := &fakeSchedulerServices{
		runs:       runs,
		workspaces: workspaces,
		events:     ch,
	}

	// construct and run scheduler
	scheduler := scheduler{
		Logger:           logr.Discard(),
		WorkspaceService: services,
		RunService:       services,
		WatchService:     services,
		queues:           make(map[string]eventHandler),
	}
	// handled chan receives events relayed to handlers
	handled := make(chan otf.Event)
	scheduler.queueFactory = &fakeQueueFactory{events: handled}
	return &scheduler, handled
}

type fakeSchedulerServices struct {
	runs       []*run.Run
	workspaces []*workspace.Workspace
	events     chan otf.Event

	WorkspaceService
	RunService
}

func (f *fakeSchedulerServices) ListRuns(context.Context, run.RunListOptions) (*run.RunList, error) {
	return &run.RunList{
		Items:      f.runs,
		Pagination: otf.NewPagination(otf.ListOptions{}, len(f.runs)),
	}, nil
}

func (f *fakeSchedulerServices) ListWorkspaces(context.Context, workspace.WorkspaceListOptions) (*workspace.WorkspaceList, error) {
	return &workspace.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(otf.ListOptions{}, len(f.workspaces)),
	}, nil
}

func (f *fakeSchedulerServices) Watch(context.Context, otf.WatchOptions) (<-chan otf.Event, error) {
	return f.events, nil
}

type fakeQueueFactory struct {
	events chan otf.Event
}

func (f *fakeQueueFactory) newQueue(queueOptions) eventHandler {
	return &fakeQueue{events: f.events}
}

type fakeQueue struct {
	events chan otf.Event
}

func (f *fakeQueue) handleEvent(ctx context.Context, event otf.Event) error {
	f.events <- event
	return nil
}
