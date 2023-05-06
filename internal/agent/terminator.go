package agent

import "sync"

// cancelable is something that is cancelable, either forcefully or gracefully.
type cancelable interface {
	cancel(force bool)
}

// terminator handles canceling items using their ID
type terminator struct {
	// mapping maps ID to cancelable item
	mapping map[string]cancelable

	mu sync.Mutex
}

func newTerminator() *terminator {
	return &terminator{
		mapping: make(map[string]cancelable),
	}
}

func (t *terminator) checkIn(id string, job cancelable) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.mapping[id] = job
}

func (t *terminator) checkOut(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.mapping, id)
}

func (t *terminator) cancel(id string, force bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if job, ok := t.mapping[id]; ok {
		job.cancel(force)
	}
}
