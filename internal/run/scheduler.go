package run

import (
	"context"
	"slices"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
	"github.com/pkg/errors"
)

// SchedulerLockID guarantees only one scheduler on a cluster is running at any
// time.
const SchedulerLockID int64 = 5577006791947779410

type (
	// scheduler performs two principle tasks :
	// (a) manages lifecycle of workspace queues, creating/destroying them
	// (b) relays run and workspace events onto queues.
	scheduler struct {
		logr.Logger

		workspaces schedulerWorkspaceClient
		runs       schedulerRunClient

		// map workspace's ID to its runs
		queues map[resource.ID]queue
	}

	queue struct {
		current *resource.ID
		backlog []resource.ID
	}

	schedulerWorkspaceClient interface {
		Watch(context.Context) (<-chan pubsub.Event[*workspace.Workspace], func())
		Unlock(context.Context, resource.ID, *resource.ID, bool) (*workspace.Workspace, error)
	}

	schedulerRunClient interface {
		List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error)
		Watch(context.Context) (<-chan pubsub.Event[*Run], func())
		EnqueuePlan(ctx context.Context, runID resource.ID) (*Run, error)
	}

	SchedulerOptions struct {
		logr.Logger

		WorkspaceClient schedulerWorkspaceClient
		RunClient       schedulerRunClient
	}
)

func NewScheduler(opts SchedulerOptions) *scheduler {
	return &scheduler{
		Logger:     opts.WithValues("component", "scheduler"),
		workspaces: opts.WorkspaceClient,
		runs:       opts.RunClient,
	}
}

// reinitialize retrieves workspaces and runs from the DB and listens to events,
// creating/deleting workspace queues accordingly and forwarding events to
// queues for scheduling.
func (s *scheduler) Start(ctx context.Context) error {
	// subscribe to workspace events
	subWorkspaces, unsubWorkspaces := s.workspaces.Watch(ctx)
	defer unsubWorkspaces()

	// subscribe to run events
	subRuns, unsubRuns := s.runs.Watch(ctx)
	defer unsubRuns()

	// Retrieve all incomplete runs
	runs, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*Run], error) {
		return s.runs.List(ctx, ListOptions{
			Statuses:    IncompleteRun,
			PageOptions: opts,
		})
	})
	if err != nil {
		return err
	}
	// ListRuns returns newest first, whereas we want oldest first.
	slices.Reverse(runs)

	// Populate queues with existing runs and make scheduling decisions
	s.queues = make(map[resource.ID]queue)
	for _, run := range runs {
		if err := s.schedule(ctx, run.WorkspaceID, run); err != nil {
			return err
		}
	}

	// Watch workspace and run events and schedule runs accordingly.
	for {
		select {
		case event, ok := <-subRuns:
			if !ok {
				return pubsub.ErrSubscriptionTerminated
			}
			if event.Type == pubsub.DeletedEvent {
				// ignore deleted run events - the only way runs are deleted is
				// if its workspace is deleted, in which case the workspace
				// queue is deleted along with any runs.
				continue
			}
			if err := s.schedule(ctx, event.Payload.WorkspaceID, event.Payload); err != nil {
				return err
			}
		case event, ok := <-subWorkspaces:
			if !ok {
				return pubsub.ErrSubscriptionTerminated
			}
			if event.Type == pubsub.DeletedEvent {
				delete(s.queues, event.Payload.ID)
				continue
			}
			// While a user has held a workspace lock runs cannot be scheduled,
			// so try to schedule again when a workspace event arrives in an
			// unlocked state.
			if !event.Payload.Locked() {
				if err := s.schedule(ctx, event.Payload.ID, nil); err != nil {
					return err
				}
			}
		}
	}
}

func (s *scheduler) schedule(ctx context.Context, workspaceID resource.ID, run *Run) error {
	if run != nil && run.PlanOnly && run.Status == RunPending {
		// Pending plan-only runs immediately have a plan enqueued and not added to the
		// workspace queue.
		_, err := s.runs.EnqueuePlan(ctx, run.ID)
		return err
	}
	q := s.queues[workspaceID]
	q, enqueue, unlock := q.process(run)
	if enqueue {
		run, err := s.runs.EnqueuePlan(ctx, *q.current)
		if err != nil {
			if errors.Is(err, workspace.ErrWorkspaceAlreadyLocked) {
				s.V(0).Info("workspace locked by user; cannot schedule run", "run", run.ID)
				return err
			}
			return err
		}
	}
	if unlock {
		_, err := s.workspaces.Unlock(ctx, workspaceID, &run.ID, false)
		if errors.Is(err, internal.ErrResourceNotFound) {
			// Workspace not found error can occur when a workspace is deleted
			// very soon after a run has completed (a quite possible scenario
			// for "ephemeral like" workspaces created and destroyed via the
			// API).
			// If this is the case there is nothing to unlock and the scheduler
			// can continue as normal.
			return nil
		} else if err != nil {
			// Any other error is treated as a transient or unexpected
			// error, so propagate the error which'll notify the user
			// via the logs and trigger the scheduler to be restarted
			// with a backoff-and-retry.
			return err
		}
	}
	s.queues[workspaceID] = q
	return nil
}

// process the queue: re-arrange the queue accordingly and determine whether a
// plan should be enqueued for a run. If run is non-nil then it is added/removed
// from the queue accordingly. Unlock is true if the workspace should be
// unlocked.
func (q queue) process(run *Run) (qq queue, enqueuePlan bool, unlock bool) {
	if run != nil {
		if q.current != nil && *q.current == run.ID {
			if run.Done() {
				q.current = nil
				// Workspace can be unlocked unless another run below is made
				// the current run.
				unlock = true
			}
		} else {
			// check if run is in backlog
			var found bool
			for i, id := range q.backlog {
				if run.ID == id {
					if run.Done() {
						// remove run from backlog
						q.backlog = append(q.backlog[:i], q.backlog[i+1:]...)
						return q, false, false
					}
					found = true
					break
				}
			}
			// add to backlog if not already in backlog
			if !found {
				q.backlog = append(q.backlog, run.ID)
			}
		}
	}
	if q.current == nil && len(q.backlog) > 0 {
		q.current = &q.backlog[0]
		q.backlog = q.backlog[1:]
		return q, true, false
	}
	return q, false, unlock
}
