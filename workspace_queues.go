package otf

import (
	"context"

	"github.com/go-logr/logr"
)

// workspaceQueues manages workspace queues of runs
type workspaceQueues struct {
	// RunService retrieves and updates runs
	RunService
	// EventService permits scheduler to subscribe to a stream of events
	EventService
	// Logger for logging various events
	logr.Logger
	// run queue for each workspace
	queues map[string][]*Run
	// queue change notifications
	changes chan struct{}
}

func (s *workspaceQueues) seed(ctx context.Context, rs RunService) error {
	runs, err := rs.List(ctx, RunListOptions{Statuses: IncompleteRunStatuses})
	if err != nil {
		return err
	}
	for _, run := range runs.Items {
		if run.IsSpeculative() {
			// speculative runs are never enqueued
			continue
		}
		s.updateQueue(run.Workspace.ID, func(q []*Run) []*Run {
			return append(q, run)
		})
	}
	return nil
}

func (s *workspaceQueues) update(run *Run) {
	if run.IsSpeculative() {
		// speculative runs are never enqueued
		return
	}
	s.updateQueue(run.Workspace.ID, func(q []*Run) []*Run {
		if i := indexRunSlice(q, run); i >= 0 {
			if run.IsDone() {
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
	})
}

// updateQueue updates a workspace queue with the return val of fn. If the queue
// doens't exist, it'll be created first. If fn changes the queue size then a
// notification is sent.
func (s *workspaceQueues) updateQueue(workspaceID string, fn func(q []*Run) []*Run) {
	q, ok := s.queues[workspaceID]
	if !ok {
		q = []*Run{}
	}
	qq := fn(q)
	if len(qq) != len(q) {
		s.changes <- struct{}{}
	}
	q = qq
}

// indexRunSlice retrieves the index of a run within a slice of runs, returning
// -1 if run is not found.
func indexRunSlice(runs []*Run, run *Run) int {
	for i, r := range runs {
		if r.ID == run.ID {
			return i
		}
	}
	return -1
}
