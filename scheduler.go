package otf

import (
	"context"

	"github.com/go-logr/logr"
)

// Scheduler is responsible for starting runs i.e. the first step, enqueuing the
// plan
type Scheduler struct {
	RunService
	incoming <-chan *Run
	WorkspaceService
	logr.Logger

	ctx context.Context
}

// NewScheduler constructs and initialises the scheduler.
func NewScheduler(ctx context.Context, logger logr.Logger, app Application) (*Scheduler, error) {
	s := &Scheduler{
		RunService:       app.RunService(),
		WorkspaceService: app.WorkspaceService(),
		Logger:           logger,
	}
	s.ctx = AddSubjectToContext(ctx, s)
	lw, err := app.RunService().ListWatch(ctx, RunListOptions{Statuses: IncompleteRun})
	if err != nil {
		return nil, err
	}
	s.incoming = lw
	return s, nil
}

func (s *Scheduler) CanAccess(Resource) bool {
	// Scheduler can access anything
	return true
}

// Start starts the scheduler daemon. Should be invoked in a go routine.
func (s *Scheduler) Start() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case run := <-s.incoming:
			if err := s.handleRun(s.ctx, run); err != nil {
				s.Error(err, "scheduling run", "run", run.ID())
			}
		}
	}
}

func (s *Scheduler) handleRun(ctx context.Context, run *Run) error {
	if run.Speculative() {
		if run.Status() == RunPending {
			// immediately enqueue plan for pending speculative runs
			_, err := s.RunService.EnqueuePlan(ctx, run.ID())
			if err != nil {
				return err
			}
		}
		// speculative runs are not enqueued so stop here
		return nil
	}
	// enqueue run and see if the run at the front of the queue needs starting.
	if err := s.WorkspaceService.UpdateQueue(run); err != nil {
		return err
	}
	queue, err := s.WorkspaceService.GetQueue(run.workspaceID)
	if err != nil {
		return err
	}
	if len(queue) > 0 && queue[0].Status() == RunPending {
		// enqueue plan for pending run at head of queue
		_, err := s.RunService.EnqueuePlan(ctx, queue[0].ID())
		if err != nil {
			return err
		}
	}
	return nil
}
