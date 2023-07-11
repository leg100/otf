// Package hooks implements the observer pattern
package hooks

import (
	"context"
	"sync"

	"github.com/leg100/otf/internal/rbac"
)

// Listener is a function that can listen and react to a hook event
type Listener func(ctx context.Context, id string) error

// Hook is a mechanism which supports the ability to dispatch data to arbitrary listener callbacks
type Hook struct {
	// action is the action that dispatches the hook
	action rbac.Action

	// before stores the functions which will be invoked before the hook action
	// occurs
	before []Listener

	// after stores the functions which will be invoked after the hook action
	// occurs
	after []Listener

	// mu stores the mutex to provide concurrency-safe operations
	mu sync.RWMutex
}

// NewHook creates a new Hook
func NewHook(action rbac.Action) *Hook {
	return &Hook{
		action: action,
		before: make([]Listener, 0),
		after:  make([]Listener, 0),
		mu:     sync.RWMutex{},
	}
}

// Dispatch invokes all listeners synchronously with the provided message
func (h *Hook) Dispatch(ctx context.Context, id string, fn func(context.Context) error) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, callback := range h.before {
		if err := callback(ctx, id); err != nil {
			return err
		}
	}

	if err := fn(ctx); err != nil {
		return err
	}

	for _, callback := range h.after {
		if err := callback(ctx, id); err != nil {
			return err
		}
	}

	return nil
}

// Before registers a callback function to be invoked before the hook action
// occurs.
func (h *Hook) Before(callback Listener) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.before = append(h.before, callback)
}

// After registers a callback function to be invoked after the hook action
// occurs.
func (h *Hook) After(callback Listener) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.after = append(h.after, callback)
}
