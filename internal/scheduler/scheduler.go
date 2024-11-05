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

		workspaces workspaceClient
		runs       runClient

		queues map[resource.ID]eventHandler
		queueFactory
	}

	workspaceClient interface {
		List(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error)
		Watch(context.Context) (<-chan pubsub.Event[*workspace.Workspace], func())
		Lock(ctx context.Context, workspaceID resource.ID, runID *resource.ID) (*workspace.Workspace, error)
		Unlock(ctx context.Context, workspaceID resource.ID, runID *resource.ID, force bool) (*workspace.Workspace, error)
		SetCurrentRun(ctx context.Context, workspaceID, runID resource.ID) (*workspace.Workspace, error)
	}

	runClient interface {
		List(ctx context.Context, opts run.ListOptions) (*resource.Page[*run.Run], error)
		Watch(context.Context) (<-chan pubsub.Event[*run.Run], func())
		EnqueuePlan(ctx context.Context, runID resource.ID) (*run.Run, error)
	}

	Options struct {
		logr.Logger

		WorkspaceClient workspaceClient
		RunClient       runClient
	}
)

func NewScheduler(opts Options) *scheduler {
	return &scheduler{
		Logger:       opts.Logger.WithValues("component", "scheduler"),
		workspaces:   opts.WorkspaceClient,
		runs:         opts.RunClient,
		queueFactory: queueMaker{},
	}
}

// reinitialize retrieves workspaces and runs from the DB and listens to events,
// creating/deleting workspace queues accordingly and forwarding events to
// queues for scheduling.
func (s *scheduler) Start(ctx context.Context) error {
	// Reset queues each time scheduler starts
	s.queues = make(map[resource.ID]eventHandler)

	// subscribe to workspace events
	subWorkspaces, unsubWorkspaces := s.workspaces.Watch(ctx)
	defer unsubWorkspaces()

	// subscribe to run events
	subRuns, unsubRuns := s.runs.Watch(ctx)
	defer unsubRuns()

	// retrieve all existing workspaces
	workspaces, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Workspace], error) {
		return s.workspaces.List(ctx, workspace.ListOptions{
			PageOptions: opts,
		})
	})
	if err != nil {
		return fmt.Errorf("retrieving existing workspaces: %w", err)
	}
	// retrieve all incomplete runs
	runs, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*run.Run], error) {
		return s.runs.List(ctx, run.ListOptions{
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
			if err := s.handleWorkspaceEvent(ctx, workspaceEvent); err != nil {
				return err
			}
		case runEvent, ok := <-runQueue:
			if !ok {
				return pubsub.ErrSubscriptionTerminated
			}
			if err := s.handleRunEvent(ctx, runEvent); err != nil {
				return err
			}
		}
	}
}

func (s *scheduler) handleWorkspaceEvent(ctx context.Context, event pubsub.Event[*workspace.Workspace]) error {
	if event.Type == pubsub.DeletedEvent {
		delete(s.queues, event.Payload.ID)
		return nil
	}
	// create workspace queue if it doesn't exist
	q, ok := s.queues[event.Payload.ID]
	if !ok {
		q = s.newQueue(queueOptions{
			Logger:          s.Logger,
			runClient:       s.runs,
			workspaceClient: s.workspaces,
			Workspace:       event.Payload,
		})
		s.queues[event.Payload.ID] = q
	}
	if err := q.handleWorkspace(ctx, event.Payload); err != nil {
		return err
	}
	return nil
}

func (s *scheduler) handleRunEvent(ctx context.Context, event pubsub.Event[*run.Run]) error {
	if event.Type == pubsub.DeletedEvent {
		// ignore deleted run events - the only way runs are deleted is
		// if its workspace is deleted, in which case the workspace
		// queue is deleted along with any runs.
		return nil
	}
	q, ok := s.queues[event.Payload.WorkspaceID]
	if !ok {
		// No queue exists for the workspace because the workspace has
		// since been deleted, which can occur when run events arrive *after*
		// the "workspace deleted" event, which is entirely possible with the
		// way the scheduler processes events.
		//
		// In this case no action need be taken on the run.
		return nil
	}
	if err := q.handleRun(ctx, event.Payload); err != nil {
		return err
	}
	return nil
}
