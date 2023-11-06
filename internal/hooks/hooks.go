// Package hooks implements the observer pattern
package hooks

import (
	"context"
	"sync"

	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

// Listener is a function that can listen and react to a hook event
type Listener[T any] func(ctx context.Context, event T) error

// Hook is a mechanism which supports the ability to dispatch data to arbitrary listener callbacks
type Hook[T any] struct {
	// db for wrapping dispatch in a transaction
	db *sql.DB

	// before stores the functions which will be invoked before the hook action
	// occurs
	before []Listener[T]

	// after stores the functions which will be invoked after the hook action
	// occurs
	after []Listener[T]

	// mu stores the mutex to provide concurrency-safe operations
	mu sync.RWMutex
}

// NewHook creates a new Hook
func NewHook[T any](db *sql.DB) *Hook[T] {
	return &Hook[T]{
		db:     db,
		before: make([]Listener[T], 0),
		after:  make([]Listener[T], 0),
	}
}

// Dispatch invokes all listeners synchronously within a transaction.
func (h *Hook[T]) Dispatch(ctx context.Context, event T, fn func(context.Context) (T, error)) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.db.Tx(ctx, func(ctx context.Context, _ pggen.Querier) error {
		for _, callback := range h.before {
			if err := callback(ctx, event); err != nil {
				return err
			}
		}

		event, err := fn(ctx)
		if err != nil {
			return err
		}

		for _, callback := range h.after {
			if err := callback(ctx, event); err != nil {
				return err
			}
		}
		return nil
	})
}

// Before registers a callback function to be invoked before the hook action
// occurs.
func (h *Hook[T]) Before(callback Listener[T]) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.before = append(h.before, callback)
}

// After registers a callback function to be invoked after the hook action
// occurs.
func (h *Hook[T]) After(callback Listener[T]) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.after = append(h.after, callback)
}
