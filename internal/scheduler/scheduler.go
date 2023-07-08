// Package scheduler is responsible for the scheduling of runs
package scheduler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

// LockID guarantees only one scheduler on a cluster is running at any
// time.
const LockID int64 = 5577006791947779410

type (
	// scheduler performs two principle tasks :
	// (a) manages lifecycle of workspace queues, creating/destroying them
	// (b) relays run and workspace events onto queues.
	scheduler struct {
		logr.Logger

		pubsub.Subscriber
		WorkspaceService
		RunService

		queues map[string]eventHandler
		queueFactory
	}

	Options struct {
		logr.Logger
		pubsub.Subscriber

		WorkspaceService
		RunService
	}

	WorkspaceService workspace.Service
	RunService       run.Service
)

func NewScheduler(opts Options) *scheduler {
	return &scheduler{
		Logger:           opts.Logger.WithValues("component", "scheduler"),
		WorkspaceService: opts.WorkspaceService,
		RunService:       opts.RunService,
		Subscriber:       opts.Subscriber,
		queueFactory:     queueMaker{},
	}
}

// reinitialize retrieves workspaces and runs from the DB and listens to events,
// creating/deleting workspace queues accordingly and forwarding events to
// queues for scheduling.
func (s *scheduler) Start(ctx context.Context) error {
	// Unsubscribe Subscribe() whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Reset queues each time scheduler starts
	s.queues = make(map[string]eventHandler)

	// subscribe to run events and workspace unlock events
	sub, err := s.Subscribe(ctx, "scheduler-")
	if err != nil {
		return err
	}

	// retrieve all existing workspaces
	workspaces, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Workspace], error) {
		return s.ListWorkspaces(ctx, workspace.ListOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		return fmt.Errorf("retrieving existing workspaces: %w", err)
	}
	// retrieve all incomplete runs
	runs, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*run.Run], error) {
		return s.ListRuns(ctx, run.RunListOptions{
			Statuses:    internal.IncompleteRun,
			PageOptions: opts,
		})
	})
	if err != nil {
		return fmt.Errorf("retrieving incomplete runs: %w", err)
	}

	// feed in existing runs and workspaces and then events to the scheduler for processing
	queue := make(chan pubsub.Event)
	go func() {
		for _, ws := range workspaces {
			queue <- pubsub.Event{
				Payload: ws,
			}
		}
		// spool existing runs in reverse order; ListRuns returns runs newest first,
		// whereas we want oldest first.
		for i := len(runs) - 1; i >= 0; i-- {
			queue <- pubsub.Event{
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
			if event.Type == pubsub.DeletedEvent {
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
				s.Error(nil, "workspace queue does not exist for run event", "workspace", payload.WorkspaceID, "run", payload.ID, "event", event.Type)
				continue
			}
			if err := q.handleEvent(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}
