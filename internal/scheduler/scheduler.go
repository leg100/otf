// Package scheduler is responsible for the scheduling of runs
package scheduler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
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

		WorkspaceService
		RunService

		queues map[string]eventHandler
		queueFactory
	}

	Options struct {
		logr.Logger

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
		queueFactory:     queueMaker{},
	}
}

// reinitialize retrieves workspaces and runs from the DB and listens to events,
// creating/deleting workspace queues accordingly and forwarding events to
// queues for scheduling.
func (s *scheduler) Start(ctx context.Context) error {
	// Reset queues each time scheduler starts
	s.queues = make(map[string]eventHandler)

	// subscribe to workspace events
	subWorkspaces, unsubWorkspaces := s.SubscribeWorkspaceEvents()
	defer unsubWorkspaces()

	// subscribe to run events
	subRuns, unsubRuns := s.SubscribeRunEvents()
	defer unsubRuns()

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
		return s.ListRuns(ctx, run.ListOptions{
			Statuses:    run.IncompleteRun,
			PageOptions: opts,
		})
	})
	if err != nil {
		return fmt.Errorf("retrieving incomplete runs: %w", err)
	}

	// feed in existing workspaces and then events to the scheduler for processing
	workspaceQueue := make(chan pubsub.Event[*workspace.Workspace])
	go func() {
		for _, ws := range workspaces {
			workspaceQueue <- pubsub.Event[*workspace.Workspace]{
				Payload: ws,
			}
		}
		for event := range subWorkspaces {
			workspaceQueue <- event
		}
		close(workspaceQueue)
	}()

	// feed in existing runs and then events to the scheduler for processing
	runQueue := make(chan pubsub.Event[*run.Run])
	go func() {
		// spool existing runs in reverse order; ListRuns returns runs newest first,
		// whereas we want oldest first.
		for i := len(runs) - 1; i >= 0; i-- {
			runQueue <- pubsub.Event[*run.Run]{
				Payload: runs[i],
			}
		}
		for event := range subRuns {
			runQueue <- event
		}
		close(runQueue)
	}()

	for {
		select {
		case workspaceEvent, ok := <-workspaceQueue:
			if !ok {
				return pubsub.ErrSubscriptionTerminated
			}
			if workspaceEvent.Type == pubsub.DeletedEvent {
				delete(s.queues, workspaceEvent.Payload.ID)
				continue
			}
			// create workspace queue if it doesn't exist
			q, ok := s.queues[workspaceEvent.Payload.ID]
			if !ok {
				q = s.newQueue(queueOptions{
					Logger:           s.Logger,
					RunService:       s.RunService,
					WorkspaceService: s.WorkspaceService,
					Workspace:        workspaceEvent.Payload,
				})
				s.queues[workspaceEvent.Payload.ID] = q
			}
			if err := q.handleWorkspace(ctx, workspaceEvent.Payload); err != nil {
				return err
			}
		case runEvent, ok := <-runQueue:
			if !ok {
				return pubsub.ErrSubscriptionTerminated
			}
			if runEvent.Type == pubsub.DeletedEvent {
				// ignore deleted run events - the only way runs are deleted is
				// if its workspace is deleted, in which case the workspace
				// queue is deleted along with any runs.
				continue
			}
			q, ok := s.queues[runEvent.Payload.WorkspaceID]
			if !ok {
				// should never happen
				s.Error(nil, "workspace queue does not exist for run event", "workspace", runEvent.Payload.WorkspaceID, "run", runEvent.Payload.ID, "event", runEvent.Type)
				continue
			}
			if err := q.handleRun(ctx, runEvent.Payload); err != nil {
				return err
			}
		}
	}
}
