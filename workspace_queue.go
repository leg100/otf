package otf

// Queue maintains a queue of runs
type Queue interface {
	Update(*Run) *Run
	Len() int
}

// WorkspaceQueue is the queue of runs for a workspace.
type WorkspaceQueue struct {
	queue []*Run
}

// Seed seeds workspace queue with a batch of runs
func (q *WorkspaceQueue) Seed(runs ...*Run) {
	for _, r := range runs {
		if r.IsSpeculative() || r.IsDone() {
			// Speculative or completed runs are never enqueued
			return
		}
		q.queue = append(q.queue, r)
	}
}

// Update updates the queue with a newly updated run.
func (q *WorkspaceQueue) Update(run *Run) {
	if run.IsSpeculative() {
		// Speculative runs are never queued
		return
	}
	if pos := q.position(run); pos >= 0 {
		if run.IsDone() {
			// Remove from queue
			q.queue = append(q.queue[:pos], q.queue[pos+1:]...)
		}
	} else {
		// Add to queue
		q.queue = append(q.queue, run)
	}
	return
}

// Startable returns a run at the head of the queue that is pending and thus can
// be started (i.e. it can be promoted to RunPlanQueued status). Otherwise nil
// is returned.
func (q *WorkspaceQueue) Startable() *Run {
	if len(q.queue) > 0 && q.queue[0].Status == RunPending {
		return q.queue[0]
	}
	return nil
}

func (q *WorkspaceQueue) Len() int {
	return len(q.queue)
}

func (q *WorkspaceQueue) position(run *Run) int {
	for i, r := range q.queue {
		if r.ID == run.ID {
			return i
		}
	}
	return -1
}
