// Package hooks implements the observer pattern
package hooks

import (
	"context"
	"sync"

	"github.com/leg100/otf/internal/rbac"
)

// Listener is a function that can listen and react to a hook event
type Listener[T any] func(ctx context.Context, event Event[T])

// Hook is a mechanism which supports the ability to dispatch data to arbitrary listener callbacks
type Hook[T any] struct {
	// action is the action that dispatches the hook
	action rbac.Action

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
func NewHook[T any](action rbac.Action) *Hook[T] {
	return &Hook[T]{
		action: action,
		before: make([]Listener[T], 0),
		after:  make([]Listener[T], 0),
		mu:     sync.RWMutex{},
	}
}

// GetAction returns the hook's action
func (h *Hook[T]) GetAction() rbac.Action {
	return h.action
}

// Before registers a callback function to be invoked after the hook action
// occurs.
func (h *Hook[T]) Before(callback Listener[T]) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.before = append(h.before, callback)
}

// After registers a callback function to be invoked before the hook action
// occurs.
func (h *Hook[T]) After(callback Listener[T]) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.after = append(h.after, callback)
}

// Dispatch invokes all listeners synchronously with the provided message
func (h *Hook[T]) Dispatch(ctx context.Context, message T, fn func(context.Context) error) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	e := newEvent[T](h, message)

	for _, callback := range h.before {
		callback(ctx, e)
	}

	if err := fn(ctx); err != nil {
		return err
	}

	for _, callback := range h.after {
		callback(ctx, e)
	}

	return nil
}
