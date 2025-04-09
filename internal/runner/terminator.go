package runner

import (
	"sync"

	"github.com/leg100/otf/internal/resource"
)

// cancelable is something that is cancelable, either forcefully or gracefully.
type cancelable interface {
	cancel(force, sendSignal bool)
}

// terminator handles canceling jobs
type terminator struct {
	// mapping maps job to a cancelable operation executing the job.
	mapping map[resource.TfeID]cancelable
	mu      sync.RWMutex
}

func (t *terminator) checkIn(jobID resource.TfeID, job cancelable) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.mapping[jobID] = job
}

func (t *terminator) checkOut(jobID resource.TfeID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.mapping, jobID)
}

func (t *terminator) cancel(jobID resource.TfeID, force, sendSignal bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if job, ok := t.mapping[jobID]; ok {
		job.cancel(force, sendSignal)
	}
}

func (t *terminator) stopAll() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, job := range t.mapping {
		job.cancel(false, false)
	}
}

func (t *terminator) totalJobs() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.mapping)
}
