package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/leg100/otf/internal/logr"
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
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		close(ch)
	})
	job1 := resource.NewTfeID(resource.JobKind)
	job2 := resource.NewTfeID(resource.JobKind)

	// subscribe to job1
	fn := signaler.awaitJobSignal(context.Background(), job1)

	// send job2 signal
	ch <- fmt.Sprintf(`{"job_id": "%s"}`, job2)
	// send job1 signal
	ch <- fmt.Sprintf(`{"job_id": "%s"}`, job1)

	// expect to receive job1 signal but not job2 signal (which would otherwise
	// block the job1 signal).
	got, err := fn()
	require.NoError(t, err)
	assert.Equal(t, jobSignal{JobID: job1}, got)

	// subscribe to job1 again
	fn = signaler.awaitJobSignal(context.Background(), job1)

	// send job1 force signal
	ch <- fmt.Sprintf(`{"job_id": "%s","force": true}`, job1)

	got, err = fn()
	require.NoError(t, err)
	assert.Equal(t, jobSignal{JobID: job1, Force: true}, got)
}
