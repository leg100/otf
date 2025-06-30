package runner

import (
	"sync"

	"github.com/leg100/otf/internal/resource"
)

// cancelable is something that is cancelable, either forcefully or gracefully.
type cancelable interface {
	cancel(force, sendSignal bool)
}

// jobTracker keeps track of a runner's active jobs.
type jobTracker struct {
	// mapping maps job to a cancelable operation executing the job.
	mapping map[resource.TfeID]cancelable
	mu      sync.RWMutex
}

func (t *jobTracker) checkIn(jobID resource.TfeID, job cancelable) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.mapping[jobID] = job
}

func (t *jobTracker) checkOut(jobID resource.TfeID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.mapping, jobID)
}

func (t *jobTracker) stopAll() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, job := range t.mapping {
		job.cancel(false, false)
	}
}

func (t *jobTracker) totalJobs() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.mapping)
}
