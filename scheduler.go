package otf

import (
	"context"

	"github.com/go-logr/logr"
)

// Scheduler is responsible for starting runs i.e. the first step, enqueuing the
// plan
type Scheduler struct {
	Application
	incoming <-chan *Run
	logr.Logger
}

// NewScheduler constructs and initialises the scheduler.
func NewScheduler(ctx context.Context, logger logr.Logger, app Application) (*Scheduler, error) {
	s := &Scheduler{
		Application: app,
		Logger:      logger,
	}
	lw, err := app.ListWatchRun(ctx, RunListOptions{Statuses: IncompleteRun})
	if err != nil {
		return nil, err
	}
	s.incoming = lw
	return s, nil
}

// Start starts the scheduler daemon. Should be invoked in a go routine.
func (s *Scheduler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case run := <-s.incoming:
			if err := s.handleRun(ctx, run); err != nil {
				s.Error(err, "scheduling run", "run", run.ID())
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
	queue, err := s.GetWorkspaceQueue(run.workspaceID)
	if err != nil {
		return err
	}
	if len(queue) > 0 && queue[0].Status() == RunPending {
		// enqueue plan for pending run at head of queue
		_, err := s.EnqueuePlan(ctx, queue[0].ID())
		if err != nil {
			return err
		}
	}
	return nil
}
