package pubsub

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/require"
)

// TestBroker_UnsubscribeDoesNotLeakGoroutine ensures that unsubscribing
// releases the goroutine watching the context, even when the context outlives
// the subscription.
//
// The server runner subscribes to job events on every iteration of its
// job-processing loop, passing a context that lives for the lifetime of the
// process. Prior to the fix, each subscription left behind a goroutine blocked
// on ctx.Done() forever, and that goroutine kept the subscription's channel -
// and its eagerly-allocated buffer of subBufferSize events - alive with it.
func TestBroker_UnsubscribeDoesNotLeakGoroutine(t *testing.T) {
	// A context that outlives every subscription, as per the server runner.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker := NewBroker[*foo](logr.Discard(), "foos")

	before := countGoroutines(t, 0)

	const subscriptions = 100
	for range subscriptions {
		_, unsub, err := broker.Subscribe(ctx)
		require.NoError(t, err)
		unsub()
	}

	after := countGoroutines(t, before)
	require.Less(t, after-before, subscriptions/2,
		"unsubscribing leaked goroutines: %d before, %d after %d subscriptions",
		before, after, subscriptions)
}

// countGoroutines waits for the goroutine count to settle at or below want,
// returning the count. Goroutines are torn down asynchronously so a bare
// runtime.NumGoroutine() is racy.
func countGoroutines(t *testing.T, want int) int {
	t.Helper()

	var n int
	for range 100 {
		n = runtime.NumGoroutine()
		if n <= want {
			return n
		}
		time.Sleep(10 * time.Millisecond)
	}
	return n
}
