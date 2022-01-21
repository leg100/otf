package otf

import "context"

type PlanEnqueuer interface {
	EnqueuePlan(ctx context.Context, id string) error
}

// Queue implementations are able to add and remove runs from a queue-like
// structure
type Queue interface {
	Add(*Run) error
	Remove(*Run) error
}

// WorkspaceQueue is the queue of runs for a workspace. The queue has at most
// one active run, which blocks other pending runs. Speculative runs do not
// block and are therefore not added to the queue.
type WorkspaceQueue struct {
	// Active is the currently active run.
	Active *Run
	// Pending is the list of pending runs waiting for the active run to
	// complete.
	Pending []*Run
	// PlanEnqueuer enqueues a plan onto the global queue
	PlanEnqueuer
}

// Add adds a run to the workspace queue.
func (q *WorkspaceQueue) Add(run *Run) error {
	// Enqueue speculative runs onto (global) queue but don't make them active
	// because they do not block pending runs
	if run.IsSpeculative() {
		return q.EnqueuePlan(context.Background(), run.ID)
	}

	// No run is current active, so make this run active
	if q.Active == nil {
		if err := q.EnqueuePlan(context.Background(), run.ID); err != nil {
			return err
		}

		q.Active = run
		return nil
	}

	// Other add run to pending queue
	q.Pending = append(q.Pending, run)

	return nil
}

// Remove removes a run from the queue.
func (q *WorkspaceQueue) Remove(run *Run) error {
	// Speculative runs are never added to the queue in the first place so they
	// do not need to be removed
	if run.IsSpeculative() {
		return nil
	}

	// Remove active run and make the first pending run the active run
	if q.Active.ID == run.ID {
		q.Active = nil
		if len(q.Pending) > 0 {
			if err := q.EnqueuePlan(context.Background(), q.Pending[0].ID); err != nil {
				return err
			}

			q.Active = q.Pending[0]
			q.Pending = q.Pending[1:]
		}
		return nil
	}

	// Remove run from pending queue
	for idx, p := range q.Pending {
		if p.ID == run.ID {
			q.Pending = append(q.Pending[:idx], q.Pending[idx+1:]...)
			return nil
		}
	}

	return nil
}
