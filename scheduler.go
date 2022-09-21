package otf

import (
	"context"

	"github.com/go-logr/logr"
)

// Scheduler is responsible for starting runs i.e. the first step, enqueuing the
// plan
type Scheduler struct {
	Application
	logr.Logger
}

// NewScheduler constructs and initialises the scheduler.
func NewScheduler(logger logr.Logger, app Application) *Scheduler {
	s := &Scheduler{
		Application: app,
		Logger:      logger.WithValues("component", "scheduler"),
	}

	return s
}

// Start starts the scheduler daemon. Should be invoked in a go routine.
func (s *Scheduler) Start(ctx context.Context) error {
	// subscribe to run events and workspace unlock events
	sub, err := s.Watch(ctx, WatchOptions{})
	if err != nil {
		s.Error(err, "creating watch")
		return err
	}

	existing, err := s.existing(ctx)
	if err != nil {
		s.Error(err, "retrieving incomplete runs")
		return err
	}

	// feed in both existing runs and events to the scheduler for processing
	queue := make(chan Event)
	go func() {
		for _, run := range existing {
			queue <- Event{Payload: run}
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
			if event.Type == EventWorkspaceUnlocked {
				ws, ok := event.Payload.(*Workspace)
				if !ok {
					s.Error(nil, "received workspace event without a workspace payload")
					continue
				}
				if err := s.checkFrontOfQueue(ctx, ws.ID()); err != nil {
					s.Error(err, "checking workspace queue", "workspace", ws.ID())
				}
			} else if run, ok := event.Payload.(*Run); ok {
				if err := s.handleRun(ctx, run); err != nil {
					s.Error(err, "handling run", "run", run.ID())
				}
			}
		}
	}
}

func (s *Scheduler) handleRun(ctx context.Context, run *Run) error {
	if run.Speculative() {
		if run.Status() == RunPending {
			// immediately enqueue plan for pending speculative runs
			_, err := s.EnqueuePlan(ctx, run.ID())
			if err != nil {
				return err
			}
		}
		// speculative runs are not enqueued so stop here
		return nil
	}
	// enqueue run and see if the run at the front of the queue needs starting.
	if err := s.UpdateWorkspaceQueue(run); err != nil {
		return err
	}
	if err := s.checkFrontOfQueue(ctx, run.WorkspaceID()); err != nil {
		return err
	}

	// Handle unlocking the workspace when a run has completed and there are no
	// more runs in the queue
	queue, err := s.GetWorkspaceQueue(run.WorkspaceID())
	if err != nil {
		return err
	}
	if run.Done() && len(queue) == 0 {
		ctx = AddSubjectToContext(ctx, run)
		_, err = s.UnlockWorkspace(ctx, WorkspaceSpec{ID: String(run.WorkspaceID())}, WorkspaceUnlockOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// checkFrontOfQueue checks the front of the workspace queue to see if there is
// a pending run to be scheduled
func (s *Scheduler) checkFrontOfQueue(ctx context.Context, workspaceID string) error {
	queue, err := s.GetWorkspaceQueue(workspaceID)
	if err != nil {
		return err
	}
	if len(queue) > 0 && queue[0].Status() == RunPending {
		// schedule run
		current, err := s.EnqueuePlan(ctx, queue[0].ID())
		if err != nil {
			return err
		}
		// propagate status change to queue
		if err := s.UpdateWorkspaceQueue(current); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) existing(ctx context.Context) ([]*Run, error) {
	// retrieve existing runs, page by page
	existing := []*Run{}
	opts := RunListOptions{Statuses: IncompleteRun}
	for {
		page, err := s.ListRuns(ctx, opts)
		if err != nil {
			return nil, err
		}
		existing = append(existing, page.Items...)
		if page.NextPage() == nil {
			break
		}
		opts.PageNumber = *page.NextPage()
	}
	// db returns runs ordered by creation date, newest first, but we want
	// oldest first, so we reverse the order
	oldest := []*Run{}
	for _, r := range existing {
		oldest = append([]*Run{r}, oldest...)
	}
	return oldest, nil
}
