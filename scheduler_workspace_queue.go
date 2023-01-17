package otf

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
)

// interfaces purely for faking purposes
type workspaceQueueFactory interface {
	NewWorkspaceQueue(app Application, logger logr.Logger, ws *Workspace) eventHandler
}

type eventHandler interface {
	handleEvent(context.Context, Event) error
}

// WorkspaceQueue schedules runs for a workspace
type WorkspaceQueue struct {
	Application
	logr.Logger

	ws      *Workspace
	current *Run
	queue   []*Run
}

type queueMaker struct{}

func (queueMaker) NewWorkspaceQueue(app Application, logger logr.Logger, ws *Workspace) eventHandler {
	return &WorkspaceQueue{
		Application: app,
		ws:          ws,
		Logger:      logger.WithValues("workspace", ws.ID()),
	}
}

func (s *WorkspaceQueue) handleEvent(ctx context.Context, event Event) error {
	switch payload := event.Payload.(type) {
	case *Workspace:
		s.ws = payload
		if event.Type == EventWorkspaceUnlocked {
			// schedule next run
			if s.current != nil {
				if err := s.scheduleRun(ctx, s.current); err != nil {
					return err
				}
			}
		}
	case *Run:
		if payload.Speculative() {
			if payload.Status() == RunPending {
				// immediately enqueue onto global queue
				_, err := s.EnqueuePlan(ctx, payload.ID())
				if err != nil {
					return err
				}
			}
		} else if s.current != nil && s.current.ID() == payload.ID() {
			// current run event; scheduler only interested if it's done
			if payload.Done() {
				// current run is done; see if there is pending run waiting to
				// take its place
				if len(s.queue) > 0 {
					if err := s.setCurrentRun(ctx, s.queue[0]); err != nil {
						return err
					}
					if err := s.scheduleRun(ctx, s.queue[0]); err != nil {
						return err
					}
					s.queue = s.queue[1:]
				} else {
					// no current run & queue is empty; unlock workspace
					s.current = nil
					// unlock workspace as run
					ctx = AddSubjectToContext(ctx, payload)
					ws, err := s.UnlockWorkspace(ctx, payload.WorkspaceID(), WorkspaceUnlockOptions{})
					if err != nil {
						return err
					}
					s.ws = ws
				}
			}
		} else if s.current == nil {
			// no current run; schedule immediately
			if err := s.setCurrentRun(ctx, payload); err != nil {
				return err
			}
			return s.scheduleRun(ctx, payload)
		} else {
			// check if run is in workspace queue
			for i, queued := range s.queue {
				if payload.ID() == queued.ID() {
					if payload.Done() {
						// remove run from queue
						s.queue = append(s.queue[:i], s.queue[i+1:]...)
						return nil
					}
				}
			}
			// run is not in queue; add it
			s.queue = append(s.queue, payload)
		}
	}
	return nil
}

func (s *WorkspaceQueue) setCurrentRun(ctx context.Context, run *Run) error {
	s.current = run

	if s.ws.latestRunID != nil && *s.ws.latestRunID == run.ID() {
		// run is already set as the workspace's latest run
		return nil
	}

	if err := s.SetCurrentRun(ctx, s.ws.ID(), run.ID()); err != nil {
		return fmt.Errorf("setting current run: %w", err)
	}
	s.ws.latestRunID = String(run.ID())

	return nil
}

func (s *WorkspaceQueue) scheduleRun(ctx context.Context, run *Run) error {
	if run.Status() != RunPending {
		// run has already been scheduled
		return nil
	}

	// if workspace is locked by a user then do not schedule;
	// instead wait for an unlock event to arrive.
	if _, locked := s.ws.GetLock().(*User); locked {
		s.V(0).Info("workspace locked by user; cannot schedule run", "run", run.ID())
		return nil
	}

	// Lock the workspace as the run
	ctx = AddSubjectToContext(ctx, run)
	ws, err := s.LockWorkspace(ctx, run.WorkspaceID(), WorkspaceLockOptions{})
	if err != nil {
		if errors.Is(err, ErrWorkspaceAlreadyLocked) {
			// User has locked workspace in the small window of time between
			// getting the lock above and attempting to enqueue plan.
			s.V(0).Info("workspace locked by user; cannot schedule run", "run", run.ID())
			return nil
		}
		return err
	}
	s.ws = ws

	// schedule the run
	current, err := s.EnqueuePlan(ctx, run.ID())
	if err != nil {
		return err
	}
	s.current = current
	return nil
}
