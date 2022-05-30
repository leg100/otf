package agent

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpooler_New tests the spooler constructor
func TestSpooler_New(t *testing.T) {
	want := otf.NewTestRun(t, "run-123", "ws-123", otf.TestRunCreateOptions{})

	spooler, err := NewSpooler(
		&testRunLister{runs: []*otf.Run{want}},
		&testSubscriber{},
		logr.Discard(),
	)
	require.NoError(t, err)

	assert.Equal(t, want, <-spooler.queue)
}

// TestSpooler_Start starts the spooler and immediately cancels it.
func TestSpooler_Start(t *testing.T) {
	spooler := &SpoolerDaemon{
		Subscriber: &testSubscriber{},
		Logger:     logr.Discard(),
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		spooler.Start(ctx)
		done <- struct{}{}
	}()

	cancel()

	<-done
}

// TestSpooler_GetRun tests retrieving a job from the spooler with a
// pre-populated queue
func TestSpooler_GetRun(t *testing.T) {
	want := otf.NewTestRun(t, "run-123", "ws-123", otf.TestRunCreateOptions{})

	spooler := &SpoolerDaemon{queue: make(chan *otf.Run, 1)}
	spooler.queue <- want

	assert.Equal(t, want, <-spooler.GetRun())
}

// TestSpooler_GetRunFromEvent tests retrieving a job from the spooler after an
// event is received
func TestSpooler_GetRunFromEvent(t *testing.T) {
	want := otf.NewTestRun(t, "run-123", "ws-123", otf.TestRunCreateOptions{})

	sub := testSubscription{c: make(chan otf.Event, 1)}

	spooler := &SpoolerDaemon{
		queue:      make(chan *otf.Run, 1),
		Subscriber: &testSubscriber{sub: sub},
		Logger:     logr.Discard(),
	}

	go spooler.Start(context.Background())

	// send event
	sub.c <- otf.Event{Type: otf.EventPlanQueued, Payload: want}

	assert.Equal(t, want, <-spooler.GetRun())
}

// TestSpooler_GetRunFromCancelation tests retrieving a job from the spooler
// after a cancelation is received
func TestSpooler_GetRunFromCancelation(t *testing.T) {
	want := otf.NewTestRun(t, "run-123", "ws-123", otf.TestRunCreateOptions{})

	sub := testSubscription{c: make(chan otf.Event, 1)}

	spooler := &SpoolerDaemon{
		cancelations: make(chan *otf.Run, 1),
		Subscriber:   &testSubscriber{sub: sub},
		Logger:       logr.Discard(),
	}

	go spooler.Start(context.Background())

	// send event
	sub.c <- otf.Event{Type: otf.EventRunCanceled, Payload: want}

	assert.Equal(t, want, <-spooler.GetCancelation())
}
