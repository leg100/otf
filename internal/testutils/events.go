package testutils

import (
	"testing"
	"time"

	"github.com/leg100/otf/internal/pubsub"
)

// Wait for an event to arrive satisfying the condition within a 10 second timeout.
func Wait[T any](t *testing.T, c <-chan pubsub.Event[T], cond func(pubsub.Event[T]) bool) {
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for event")
		case event := <-c:
			if cond(event) {
				return
			}
		}
	}
}
