package scheduler

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// queue enqueues and schedules runs for a workspace
	queue struct {
		logr.Logger

		WorkspaceService
		RunService

		ws      *workspace.Workspace
		current *run.Run
		queue   []*run.Run
	}

	queueOptions struct {
		logr.Logger

		WorkspaceService
		RunService

		*workspace.Workspace
	}

	queueMaker struct{}
)

func (queueMaker) newQueue(opts queueOptions) eventHandler {
	return &queue{
		Logger:           opts.WithValues("workspace", opts.Workspace.ID),
		RunService:       opts.RunService,
		WorkspaceService: opts.WorkspaceService,
		ws:               opts.Workspace,
	}
}

func (q *queue) handleEvent(ctx context.Context, event pubsub.Event) error {
	switch payload := event.Payload.(type) {
	case *workspace.Workspace:
		q.ws = payload
		// workspace state has changed; pessimistically schedule the current run
		// in case the workspace has been unlocked.
		if q.current != nil {
			if err := q.scheduleRun(ctx, q.current); err != nil {
				return err
			}
		}
	case *run.Run:
		if payload.PlanOnly {
			if payload.Status == internal.RunPending {
				// immediately enqueue onto global queue
				_, err := q.EnqueuePlan(ctx, payload.ID)
				if err != nil {
					return err
				}
			}
		} else if q.current != nil && q.current.ID == payload.ID {
			// current run event; scheduler only interested if it's done
			if payload.Done() {
				// current run is done; see if there is pending run waiting to
				// take its place
				if len(q.queue) > 0 {
					if err := q.setCurrentRun(ctx, q.queue[0]); err != nil {
						return err
					}
					if err := q.scheduleRun(ctx, q.queue[0]); err != nil {
						return err
					}
					q.queue = q.queue[1:]
				} else {
					// no current run & queue is empty; unlock workspace
					q.current = nil
					ws, err := q.UnlockWorkspace(ctx, q.ws.ID, &payload.ID, false)
					if err != nil {
						return err
					}
					q.ws = ws
				}
			}
		} else if q.current == nil {
			// no current run; schedule immediately
			if err := q.setCurrentRun(ctx, payload); err != nil {
				return err
			}
			return q.scheduleRun(ctx, payload)
		} else {
			// check if run is in workspace queue
			for i, queued := range q.queue {
				if payload.ID == queued.ID {
					if payload.Done() {
						// remove run from queue
						q.queue = append(q.queue[:i], q.queue[i+1:]...)
						return nil
					}
				}
			}
			// run is not in queue; add it
			q.queue = append(q.queue, payload)
		}
	}
	return nil
}

func (q *queue) setCurrentRun(ctx context.Context, run *run.Run) error {
	q.current = run

	if q.ws.LatestRun != nil && q.ws.LatestRun.ID == run.ID {
		// run is already set as the workspace's latest run
		return nil
	}

	ws, err := q.SetCurrentRun(ctx, q.ws.ID, run.ID)
	if err != nil {
		return fmt.Errorf("setting current run: %w", err)
	}
	q.ws = ws

	return nil
}

func (q *queue) scheduleRun(ctx context.Context, run *run.Run) error {
	if run.Status != internal.RunPending {
		// run has already been scheduled
		return nil
	}

	// if workspace is userLocked by a user then do not schedule;
	// instead wait for an unlock event to arrive.
	if q.ws.Lock != nil && q.ws.Lock.LockKind == workspace.UserLock {
		q.V(0).Info("workspace locked by user; cannot schedule run", "run", run.ID)
		return nil
	}

	ws, err := q.LockWorkspace(ctx, q.ws.ID, &run.ID)
	if err != nil {
		if errors.Is(err, internal.ErrWorkspaceAlreadyLocked) {
			// User has locked workspace in the small window of time between
			// getting the lock above and attempting to enqueue plan.
			q.V(0).Info("workspace locked by user; cannot schedule run", "run", run.ID)
			return nil
		}
		return err
	}
	q.ws = ws

	// schedule the run
	current, err := q.EnqueuePlan(ctx, run.ID)
	if err != nil {
		return err
	}
	q.current = current
	return nil
}
