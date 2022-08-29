package agent

import "sync"

// Cancelable is something that is cancelable, either forcefully or gracefully.
type Cancelable interface {
	Cancel(force bool) error
}

// Terminator handles canceling items using their ID
type Terminator struct {
	// mapping maps ID to cancelable item
	mapping map[string]Cancelable

	mu sync.Mutex
}

func NewTerminator() *Terminator {
	return &Terminator{
		mapping: make(map[string]Cancelable),
	}
}

func (t *Terminator) CheckIn(id string, job Cancelable) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.mapping[id] = job
}

func (t *Terminator) CheckOut(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.mapping, id)
}

func (t *Terminator) Cancel(id string, force bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if job, ok := t.mapping[id]; ok {
		job.Cancel(force)
	}
}
