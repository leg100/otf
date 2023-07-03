package scheduler

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

func newTestScheduler(workspaces []*workspace.Workspace, runs []*run.Run, events ...pubsub.Event) (*scheduler, <-chan pubsub.Event) {
	ch := make(chan pubsub.Event, len(events))
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
		Subscriber:       services,
		queues:           make(map[string]eventHandler),
	}
	// handled chan receives events relayed to handlers
	handled := make(chan pubsub.Event)
	scheduler.queueFactory = &fakeQueueFactory{events: handled}
	return &scheduler, handled
}

type fakeSchedulerServices struct {
	runs       []*run.Run
	workspaces []*workspace.Workspace
	events     chan pubsub.Event

	WorkspaceService
	RunService
}

func (f *fakeSchedulerServices) ListRuns(context.Context, run.RunListOptions) (*run.RunList, error) {
	return &run.RunList{
		Items:      f.runs,
		Pagination: resource.NewPagination(resource.ListOptions{}, int64(len(f.runs))),
	}, nil
}

func (f *fakeSchedulerServices) ListWorkspaces(context.Context, workspace.ListOptions) (*workspace.WorkspaceList, error) {
	return &workspace.WorkspaceList{
		Items:      f.workspaces,
		Pagination: resource.NewPagination(resource.ListOptions{}, int64(len(f.workspaces))),
	}, nil
}

func (f *fakeSchedulerServices) Subscribe(context.Context, string) (<-chan pubsub.Event, error) {
	return f.events, nil
}

type fakeQueueFactory struct {
	events chan pubsub.Event
}

func (f *fakeQueueFactory) newQueue(queueOptions) eventHandler {
	return &fakeQueue{events: f.events}
}

type fakeQueue struct {
	events chan pubsub.Event
}

func (f *fakeQueue) handleEvent(ctx context.Context, event pubsub.Event) error {
	f.events <- event
	return nil
}
