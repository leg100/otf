package scheduler

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/workspace"
)

// queue enqueues and schedules runs for a workspace
type queue struct {
	otf.Application
	logr.Logger

	ws      workspace.Workspace
	current *run.Run
	queue   []*run.Run
}

type queueMaker struct{}

func (queueMaker) newQueue(app otf.Application, logger logr.Logger, ws workspace.Workspace) eventHandler {
	return &queue{
		Application: app,
		ws:          ws,
		Logger:      logger.WithValues("workspace", ws.ID),
	}
}

func (q *queue) handleEvent(ctx context.Context, event otf.Event) error {
	switch payload := event.Payload.(type) {
	case workspace.Workspace:
		q.ws = payload
		if event.Type == workspace.EventUnlocked {
			if q.current != nil {
				if err := q.scheduleRun(ctx, q.current); err != nil {
					return err
				}
			}
		}
	case run.Run:
		if payload.Speculative() {
			if payload.Status() == run.RunPending {
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
					// unlock workspace as run
					ctx = otf.AddSubjectToContext(ctx, payload)
					ws, err := q.UnlockWorkspace(ctx, q.ws.ID, workspace.WorkspaceUnlockOptions{})
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

	if q.ws.LatestRunID != nil && *q.ws.LatestRunID == run.ID {
		// run is already set as the workspace's latest run
		return nil
	}

	ws, err := q.SetCurrentRun(ctx, q.ws.ID, run.ID())
	if err != nil {
		return fmt.Errorf("setting current run: %w", err)
	}
	q.ws = ws

	return nil
}

func (q *queue) scheduleRun(ctx context.Context, run *run.Run) error {
	if run.Status() != run.RunPending {
		// run has already been scheduled
		return nil
	}

	// if workspace is locked by a user then do not schedule;
	// instead wait for an unlock event to arrive.
	if _, locked := q.ws.LockState().(*otf.User); locked {
		q.V(0).Info("workspace locked by user; cannot schedule run", "run", run.ID())
		return nil
	}

	// Lock the workspace as the run
	ws, err := q.LockWorkspace(otf.AddSubjectToContext(ctx, run), q.ws.ID, workspace.WorkspaceLockOptions{})
	if err != nil {
		if errors.Is(err, otf.ErrWorkspaceAlreadyLocked) {
			// User has locked workspace in the small window of time between
			// getting the lock above and attempting to enqueue plan.
			q.V(0).Info("workspace locked by user; cannot schedule run", "run", run.ID())
			return nil
		}
		return err
	}
	q.ws = ws

	// schedule the run
	current, err := q.EnqueuePlan(ctx, run.ID())
	if err != nil {
		return err
	}
	q.current = current
	return nil
}
