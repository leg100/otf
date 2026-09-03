package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobSignaler(t *testing.T) {
	// Setup signaler and relay and cleanup them up at end of test ensuring the
	// relay terminates cleanly
	signaler := newJobSignaler(logr.Discard(), nil)
	ch := make(chan string)
	go func() {
		err := signaler.relay(ch)
		require.Equal(t, pubsub.ErrSubscriptionTerminated, err)
	}()
	t.Cleanup(func() {
		close(ch)
	})
	job1 := resource.NewTfeID(resource.JobKind)
	job2 := resource.NewTfeID(resource.JobKind)

	// subscribe to job1
	fn := signaler.awaitJobSignal(t.Context(), job1)

	// send job2 signal
	ch <- fmt.Sprintf(`{"job_id": "%s"}`, job2)
	// send job1 signal
	ch <- fmt.Sprintf(`{"job_id": "%s"}`, job1)

	// expect to receive job1 signal but not job2 signal (which would otherwise
	// block the job1 signal).
	got, err := fn()
	require.NoError(t, err)
	assert.Equal(t, JobSignal{JobID: job1}, got)

	// subscribe to job1 again
	fn = signaler.awaitJobSignal(t.Context(), job1)

	// send job1 force signal
	ch <- fmt.Sprintf(`{"job_id": "%s","force": true}`, job1)

	got, err = fn()
	require.NoError(t, err)
	assert.Equal(t, JobSignal{JobID: job1, Force: true}, got)
}

// TestJobSignaler_ResubscribeSameJob checks that a job can be subscribed to more
// than once concurrently. An operation re-establishes its subscription whenever
// the long-lived request is interrupted, e.g. by a proxy timing it out, and the
// new subscription can be established before the old one is torn down.
func TestJobSignaler_ResubscribeSameJob(t *testing.T) {
	signaler := newJobSignaler(logr.Discard(), nil)
	ch := make(chan string)
	go func() {
		err := signaler.relay(ch)
		require.Equal(t, pubsub.ErrSubscriptionTerminated, err)
	}()
	t.Cleanup(func() {
		close(ch)
	})
	job := resource.NewTfeID(resource.JobKind)

	// The first subscription's context is canceled, as happens when the client
	// gives up on the interrupted request.
	firstCtx, cancelFirst := context.WithCancel(t.Context())
	first := signaler.awaitJobSignal(firstCtx, job)

	// The replacement subscription is established before the first is torn down.
	second := signaler.awaitJobSignal(t.Context(), job)

	cancelFirst()
	_, err := first()
	require.ErrorIs(t, err, context.Canceled)

	// The surviving subscription must still receive the signal, rather than
	// having been closed on the first subscription's behalf.
	ch <- fmt.Sprintf(`{"job_id": "%s","force": true}`, job)

	got, err := second()
	require.NoError(t, err)
	assert.Equal(t, JobSignal{JobID: job, Force: true}, got)

	// No subscriptions are left behind.
	signaler.mu.Lock()
	defer signaler.mu.Unlock()
	assert.Empty(t, signaler.subscribers)
}
