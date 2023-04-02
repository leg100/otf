// Package scheduler is responsible for the scheduling of runs
package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/workspace"
	"gopkg.in/cenkalti/backoff.v1"
)

// lockID guarantees only one scheduler on a cluster is running at any
// time.
const lockID int64 = 5577006791947779410

type (
	// scheduler performs two principle tasks :
	// (a) manages lifecycle of workspace queues, creating/destroying them
	// (b) relays run and workspace events onto queues.
	scheduler struct {
		logr.Logger

		otf.Subscriber
		WorkspaceService
		RunService

		queues map[string]eventHandler
		queueFactory
	}

	Options struct {
		logr.Logger
		otf.DB
		otf.Subscriber

		WorkspaceService
		RunService
	}

	WorkspaceService workspace.Service
	RunService       run.Service
)

// Start constructs and initialises the scheduler.
// start starts the scheduler daemon. Should be invoked in a go routine.
func Start(ctx context.Context, opts Options) error {
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "scheduler"})

	sched := &scheduler{
		Logger:           opts.Logger.WithValues("component", "scheduler"),
		WorkspaceService: opts.WorkspaceService,
		RunService:       opts.RunService,
		Subscriber:       opts.Subscriber,
		queues:           make(map[string]eventHandler),
		queueFactory:     queueMaker{},
	}
	sched.V(2).Info("started")

	op := func() error {
		// block on getting an exclusive lock
		err := opts.WaitAndLock(ctx, lockID, func() error {
			return sched.reinitialize(ctx)
		})
		if ctx.Err() != nil {
			return nil // exit
		}
		return err // retry
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
		sched.Error(err, "restarting scheduler")
	})
}

// reinitialize retrieves workspaces and runs from the DB and listens to events,
// creating/deleting workspace queues accordingly and forwarding events to
// queues for scheduling.
func (s *scheduler) reinitialize(ctx context.Context) error {
	// Unsubscribe Subscribe() whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to run events and workspace unlock events
	sub, err := s.Subscribe(ctx, "scheduler-")
	if err != nil {
		return err
	}

	// retrieve existing workspaces, page by page
	workspaces := []*workspace.Workspace{}
	workspaceListOpts := workspace.ListOptions{
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
	// feed in existing runs and workspaces and then events to the scheduler for processing
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
		close(queue)
	}()

	for event := range queue {
		switch payload := event.Payload.(type) {
		case *workspace.Workspace:
			if event.Type == otf.EventWorkspaceDeleted {
				delete(s.queues, payload.ID)
				continue
			}
			// create workspace queue if it doesn't exist
			q, ok := s.queues[payload.ID]
			if !ok {
				q = s.newQueue(queueOptions{
					Logger:           s.Logger,
					RunService:       s.RunService,
					WorkspaceService: s.WorkspaceService,
					Workspace:        payload,
				})
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
	return nil
}
