// Package scheduler is responsible for the scheduling of runs
package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"gopkg.in/cenkalti/backoff.v1"
)

// scheduler performs two principle tasks :
// (a) manages lifecycle of workspace queues, creating/destroying them
// (b) relays run and workspace events onto queues.
type scheduler struct {
	otf.Application
	logr.Logger
	queues map[string]eventHandler
	queueFactory
}

// newScheduler constructs and initialises the scheduler.
func newScheduler(logger logr.Logger, app otf.Application) *scheduler {
	s := &scheduler{
		Application:  app,
		Logger:       logger.WithValues("component", "scheduler"),
		queues:       make(map[string]eventHandler),
		queueFactory: queueMaker{},
	}
	s.V(2).Info("started")

	return s
}

// start starts the scheduler daemon. Should be invoked in a go routine.
func (s *scheduler) start(ctx context.Context) error {
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{"scheduler"})

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
func (s *scheduler) reinitialize(ctx context.Context) error {
	// Unsubscribe Watch() whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to run events and workspace unlock events
	sub, err := s.Watch(ctx, otf.WatchOptions{Name: otf.String("scheduler")})
	if err != nil {
		return err
	}

	// retrieve existing workspaces, page by page
	workspaces := []*workspace.Workspace{}
	workspaceListOpts := workspace.WorkspaceListOptions{
		ListOptions: otf.ListOptions{PageSize: otf.MaxPageSize},
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
	runs := []*run.Run{}
	runListOpts := run.RunListOptions{
		Statuses:    otf.IncompleteRun,
		ListOptions: otf.ListOptions{PageSize: otf.MaxPageSize},
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
	queue := make(chan otf.Event)
	go func() {
		for _, ws := range workspaces {
			queue <- otf.Event{
				Type:    otf.EventWorkspaceCreated,
				Payload: ws,
			}
		}
		// spool existing runs in reverse order; ListRuns returns runs newest first,
		// whereas we want oldest first.
		for i := len(runs) - 1; i >= 0; i-- {
			queue <- otf.Event{
				Type:    otf.EventRunStatusUpdate,
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
			case *workspace.Workspace:
				if event.Type == otf.EventWorkspaceDeleted {
					delete(s.queues, payload.ID)
					continue
				}
				// create workspace queue if it doesn't exist
				q, ok := s.queues[payload.ID]
				if !ok {
					q = s.newQueue(s.Application, s.Logger, payload)
					s.queues[payload.ID] = q
				}
				if err := q.handleEvent(ctx, event); err != nil {
					return err
				}
			case *run.Run:
				q, ok := s.queues[payload.WorkspaceID]
				if !ok {
					// should never happen
					s.Error(fmt.Errorf("workspace queue does not exist for run event"), "workspace", payload.WorkspaceID, "run", payload.ID)
					continue
				}
				if err := q.handleEvent(ctx, event); err != nil {
					return err
				}
			}
		}
	}
}
