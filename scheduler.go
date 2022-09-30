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
	// TODO: rename to queues
	queues map[string]eventHandler
	workspaceQueueFactory
}

// NewScheduler constructs and initialises the scheduler.
func NewScheduler(logger logr.Logger, app Application) *Scheduler {
	s := &Scheduler{
		Application:           app,
		Logger:                logger.WithValues("component", "scheduler"),
		queues:            make(map[string]eventHandler),
		workspaceQueueFactory: queueMaker{},
	}

	return s
}

// Start starts the scheduler daemon. Should be invoked in a go routine.
func (s *Scheduler) Start(ctx context.Context) error {
	op := func() error {
		return s.reinitialize(ctx)
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
		s.Error(err, "reinitializing scheduler")
	})
}

// reinitialize pulls incomplete runs from the DB and listens to run events,
// passing them onto workspace queues.
func (s *Scheduler) reinitialize(ctx context.Context) error {
	// subscribe to run events and workspace unlock events
	sub, err := s.Watch(ctx, WatchOptions{})
	if err != nil {
		return err
	}

	// retrieve existing incomplete runs, page by page
	existing := []*Run{}
	opts := RunListOptions{Statuses: IncompleteRun}
	for {
		page, err := s.ListRuns(ctx, opts)
		if err != nil {
			return fmt.Errorf("retrieving incomplete runs: %w", err)
		}
		existing = append(existing, page.Items...)
		if page.NextPage() == nil {
			break
		}
		opts.PageNumber = *page.NextPage()
	}
	// feed in both existing runs and events to the scheduler for processing
	queue := make(chan Event)
	go func() {
		// spool existing runs in reverse order; ListRuns returns runs newest first,
		// whereas we want oldest first.
		for i := len(existing) - 1; i >= 0; i-- {
			queue <- Event{
				Type:    EventRunStatusUpdate,
				Payload: existing[i],
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
				// create workspace queue if it doesn't exist
				q, ok := s.queues[payload.ID()]
				if !ok {
					q = s.createQueue(ctx, payload)
				}
				if err := q.handleEvent(ctx, event); err != nil {
					return err
				}
			case *Run:
				// create workspace queue if it doesn't exist
				q, ok := s.queues[payload.WorkspaceID()]
				if !ok {
					ws, err := s.GetWorkspace(ctx, WorkspaceSpec{ID: String(payload.WorkspaceID())})
					if err != nil {
						return err
					}
					q = s.createQueue(ctx, ws)
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
