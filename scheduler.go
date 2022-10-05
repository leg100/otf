package otf

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"gopkg.in/cenkalti/backoff.v1"
)

// Scheduler enqueues pending runs onto workspace queues and current runs onto
// the global queue (for processing by agents).
type Scheduler struct {
	Application
	logr.Logger
	queues map[string]eventHandler
	workspaceQueueFactory
}

// NewScheduler constructs and initialises the scheduler.
func NewScheduler(logger logr.Logger, app Application) *Scheduler {
	s := &Scheduler{
		Application:           app,
		Logger:                logger.WithValues("component", "scheduler"),
		queues:                make(map[string]eventHandler),
		workspaceQueueFactory: queueMaker{},
	}
	s.Info("started")

	return s
}

// Start starts the scheduler daemon. Should be invoked in a go routine.
func (s *Scheduler) Start(ctx context.Context) error {
	op := func() error {
		return s.reinitialize(ctx)
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
		s.Error(err, "restarting scheduler")
	})
}

// reinitialize retrieves workspaces and runs from the DB and listens to events,
// creating/deleting workspace queues accordingly and forwarding events to
// queues for scheduling.
func (s *Scheduler) reinitialize(ctx context.Context) error {
	// Unsubscribe Watch() whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to run events and workspace unlock events
	sub, err := s.Watch(ctx, WatchOptions{Name: String("scheduler")})
	if err != nil {
		return err
	}

	// retrieve existing workspaces, page by page
	workspaces := []*Workspace{}
	workspaceListOpts := WorkspaceListOptions{
		ListOptions: ListOptions{PageSize: MaxPageSize},
	}
	for {
		page, err := s.ListWorkspaces(ctx, workspaceListOpts)
		if err != nil {
			return fmt.Errorf("retrieving existing workspaces: %w", err)
		}
		workspaces = append(workspaces, page.Items...)
		if page.NextPage() == nil {
			break
		}
		workspaceListOpts.PageNumber = *page.NextPage()
	}
	// retrieve runs incomplete runs, page by page
	runs := []*Run{}
	runListOpts := RunListOptions{
		Statuses:    IncompleteRun,
		ListOptions: ListOptions{PageSize: MaxPageSize},
	}
	for {
		page, err := s.ListRuns(ctx, runListOpts)
		if err != nil {
			return fmt.Errorf("retrieving incomplete runs: %w", err)
		}
		runs = append(runs, page.Items...)
		if page.NextPage() == nil {
			break
		}
		runListOpts.PageNumber = *page.NextPage()
	}
	// feed in existing objects and then events to the scheduler for processing
	queue := make(chan Event)
	go func() {
		for _, ws := range workspaces {
			queue <- Event{
				Type:    EventWorkspaceCreated,
				Payload: ws,
			}
		}
		// spool existing runs in reverse order; ListRuns returns runs newest first,
		// whereas we want oldest first.
		for i := len(runs) - 1; i >= 0; i-- {
			queue <- Event{
				Type:    EventRunStatusUpdate,
				Payload: runs[i],
			}
		}
		for event := range sub {
			queue <- event
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-queue:
			switch payload := event.Payload.(type) {
			case *Workspace:
				if event.Type == EventWorkspaceDeleted {
					delete(s.queues, payload.ID())
					continue
				}
				// create workspace queue if it doesn't exist
				q, ok := s.queues[payload.ID()]
				if !ok {
					q = s.createQueue(ctx, payload)
				}
				if err := q.handleEvent(ctx, event); err != nil {
					return err
				}
			case *Run:
				q, ok := s.queues[payload.WorkspaceID()]
				if !ok {
					// should never happen
					s.Error(fmt.Errorf("workspace queue does not exist for run event"), "workspace", payload.WorkspaceID(), "run", payload.ID())
					continue
				}
				if err := q.handleEvent(ctx, event); err != nil {
					return err
				}
			}
		}
	}
}

func (s *Scheduler) createQueue(ctx context.Context, ws *Workspace) eventHandler {
	q := s.NewWorkspaceQueue(s.Application, s.Logger, ws)
	s.queues[ws.ID()] = q
	return q
}
