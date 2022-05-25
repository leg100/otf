package otf

import "context"

// workspaceQueue is a queue of runs for a workspace
type workspaceQueue []*Run

// update queue with a run
func (q workspaceQueue) update(run *Run) workspaceQueue {
	if i := indexRunSlice(q, run); i >= 0 {
		if run.Done() {
			// remove run from queue
			q = append(q[:i], q[i+1:]...)
		} else {
			// update in-place
			q[i] = run
		}
	} else {
		// add run to end of queue
		q = append(q, run)
	}
	return q
}

// start run at the front of the queue if not already started.
func (q workspaceQueue) startRun(ctx context.Context, rs RunService) (workspaceQueue, error) {
	if len(q) > 0 && q[0].Status() == RunPending {
		run, err := rs.Start(ctx, q[0].ID())
		if err != nil {
			return q, err
		}
		q[0] = run
		return q, nil
	}
	return q, nil
}

// indexRunSlice retrieves the index of a run within a slice of runs, returning
// -1 if not found.
func indexRunSlice(runs []*Run, run *Run) int {
	for i, r := range runs {
		if r.ID() == run.ID() {
			return i
		}
	}
	return -1
}
