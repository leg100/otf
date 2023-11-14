package agent

import "sync"

// cancelable is something that is cancelable, either forcefully or gracefully.
type cancelable interface {
	cancel(force bool)
}

// terminator handles canceling jobs
type terminator struct {
	// mapping maps job to a cancelable worker carrying out the job.
	mapping map[JobSpec]cancelable
	mu      sync.RWMutex
}

func (t *terminator) checkIn(spec JobSpec, job cancelable) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.mapping[spec] = job
}

func (t *terminator) checkOut(spec JobSpec) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.mapping, spec)
}

func (t *terminator) cancel(spec JobSpec, force bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if job, ok := t.mapping[spec]; ok {
		job.cancel(force)
	}
}

func (t *terminator) totalJobs() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.mapping)
}
