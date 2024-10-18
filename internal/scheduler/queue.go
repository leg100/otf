package scheduler

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// queue enqueues and schedules runs for a workspace
	queue struct {
		logr.Logger

		workspaceClient
		runClient

		ws      *workspace.Workspace
		current *otfrun.Run
		queue   []*otfrun.Run
	}

	queueOptions struct {
		logr.Logger

		workspaceClient
		runClient

		*workspace.Workspace
	}

	queueMaker struct{}
)

func (queueMaker) newQueue(opts queueOptions) eventHandler {
	return &queue{
		Logger:          opts.WithValues("workspace", opts.Workspace.ID),
		runClient:       opts.runClient,
		workspaceClient: opts.workspaceClient,
		ws:              opts.Workspace,
	}
}

func (q *queue) handleWorkspace(ctx context.Context, workspace *workspace.Workspace) error {
	q.ws = workspace
	// workspace state has changed; pessimistically schedule the current run
	// in case the workspace has been unlocked.
	if q.current != nil {
		if err := q.scheduleRun(ctx, q.current); err != nil {
			return err
		}
	}
	return nil
}

func (q *queue) handleRun(ctx context.Context, run *otfrun.Run) error {
	if run.PlanOnly {
		if run.Status == otfrun.RunPending {
			// immediately enqueue onto global queue
			_, err := q.EnqueuePlan(ctx, run.ID)
			if err != nil {
				return err
			}
		}
	} else if q.current != nil && q.current.ID == run.ID {
		// current run event; scheduler only interested if it's done
		if run.Done() {
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
				ws, err := q.Unlock(ctx, q.ws.ID, &run.ID, false)
				if errors.Is(err, internal.ErrResourceNotFound) {
					// Workspace not found occurs when the workspace has been
					// deleted and the scheduler hasn't yet processed the
					// corresponding "workspace deleted" event. In which there
					// is nothing to unlock and the scheduler can continue as
					// normal.
					return nil
				} else if err != nil {
					// Any other error is treated as a transient or unexpected
					// error, so propagate the error which'll notify the user
					// via the logs and trigger the scheduler to be restarted
					// with a backoff-and-retry.
					return err
				}
				q.ws = ws
			}
		}
	} else if q.current == nil {
		// no current run; schedule immediately
		if err := q.setCurrentRun(ctx, run); err != nil {
			return err
		}
		return q.scheduleRun(ctx, run)
	} else {
		// check if run is in workspace queue
		for i, queued := range q.queue {
			if run.ID == queued.ID {
				if run.Done() {
					// remove run from queue
					q.queue = append(q.queue[:i], q.queue[i+1:]...)
					return nil
				}
			}
		}
		// run is not in queue; add it
		q.queue = append(q.queue, run)
	}
	return nil
}

func (q *queue) setCurrentRun(ctx context.Context, run *otfrun.Run) error {
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

func (q *queue) scheduleRun(ctx context.Context, run *otfrun.Run) error {
	if run.Status != otfrun.RunPending {
		// run has already been scheduled
		return nil
	}

	// if workspace is userLocked by a user then do not schedule;
	// instead wait for an unlock event to arrive.
	if q.ws.Lock != nil && q.ws.Lock.LockKind == workspace.UserLock {
		q.V(0).Info("workspace locked by user; cannot schedule run", "run", run.ID)
		return nil
	}

	ws, err := q.Lock(ctx, q.ws.ID, &run.ID)
	if err != nil {
		if errors.Is(err, workspace.ErrWorkspaceAlreadyLocked) {
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
