package inmem

import (
	"github.com/go-logr/logr"
	tfe "github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

type WorkspaceQueue struct {
	active  *ots.Run
	pending []*ots.Run
	// RunService retrieves and updates runs
	ots.RunService
	logr.Logger
}

type WorkspaceQueueOption func(*WorkspaceQueue)

func NewWorkspaceQueue(rs ots.RunService, logger logr.Logger, workspaceID string, opts ...WorkspaceQueueOption) *WorkspaceQueue {
	q := &WorkspaceQueue{
		RunService: rs,
		Logger:     logger.WithValues("workspace", workspaceID),
	}

	for _, o := range opts {
		o(q)
	}

	return q
}

func WithActive(run *ots.Run) WorkspaceQueueOption {
	return func(q *WorkspaceQueue) {
		q.active = run
	}
}

func WithPending(runs []*ots.Run) WorkspaceQueueOption {
	return func(q *WorkspaceQueue) {
		q.pending = runs
	}
}

func (q *WorkspaceQueue) addRun(run *ots.Run) {
	// Enqueue speculative runs but don't make them active because they do not
	// block pending runs
	if run.IsSpeculative() {
		_, err := q.UpdateStatus(run.ID, tfe.RunPlanQueued)
		if err != nil {
			q.Error(err, "unable to enqueue run", "run", run.ID)
		}

		return
	}

	// No run is current active, so make this run active
	if q.active == nil {
		_, err := q.UpdateStatus(run.ID, tfe.RunPlanQueued)
		if err != nil {
			q.Error(err, "unable to enqueue run", "run", run.ID)
		}

		q.active = run
		return
	}

	// Other add run to pending queue
	q.pending = append(q.pending, run)
}

func (q *WorkspaceQueue) removeRun(run *ots.Run) {
	// Speculative runs are never added to the queue in the first place so they
	// do not need to be removed
	if run.IsSpeculative() {
		return
	}

	// Remove active run and make the first pending run the active run
	if q.active.ID == run.ID {
		q.active = nil
		if len(q.pending) > 0 {
			_, err := q.UpdateStatus(q.pending[0].ID, tfe.RunPlanQueued)
			if err != nil {
				q.Error(err, "unable to enqueue run", "run", run.ID)
			}

			q.active = q.pending[0]
			q.pending = q.pending[1:]
		}
		return
	}

	// Remove run from pending queue
	for idx, p := range q.pending {
		if p.ID == run.ID {
			q.pending = append(q.pending[:idx], q.pending[idx+1:]...)
			return
		}
	}
}
