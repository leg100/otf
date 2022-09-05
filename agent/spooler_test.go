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
	want := otf.NewTestRun(t, otf.TestRunCreateOptions{Status: otf.RunPlanQueued})

	spooler, err := NewSpooler(
		&testRunService{runs: []*otf.Run{want}},
		&testWatcher{},
		logr.Discard(),
	)
	require.NoError(t, err)

	assert.Equal(t, want, <-spooler.queue)
}

// TestSpooler_Start starts the spooler and immediately cancels it.
func TestSpooler_Start(t *testing.T) {
	spooler := &SpoolerDaemon{
		Watcher: &testWatcher{},
		Logger:  logr.Discard(),
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
	want := otf.NewTestRun(t, otf.TestRunCreateOptions{})

	spooler := &SpoolerDaemon{queue: make(chan *otf.Run, 1)}
	spooler.queue <- want

	assert.Equal(t, want, <-spooler.GetRun())
}

// TestSpooler_GetRunFromEvent tests retrieving a job from the spooler after an
// event is received
func TestSpooler_GetRunFromEvent(t *testing.T) {
	want := otf.NewTestRun(t, otf.TestRunCreateOptions{Status: otf.RunPlanQueued})

	ch := make(chan otf.Event, 1)

	spooler := &SpoolerDaemon{
		queue:   make(chan *otf.Run, 1),
		Watcher: &testWatcher{ch: ch},
		Logger:  logr.Discard(),
	}

	go spooler.Start(context.Background())

	// send event
	ch <- otf.Event{Type: otf.EventRunStatusUpdate, Payload: want}

	assert.Equal(t, want, <-spooler.GetRun())
}

// TestSpooler_GetRunFromCancelation tests retrieving a job from the spooler
// after a cancelation is received
func TestSpooler_GetRunFromCancelation(t *testing.T) {
	want := otf.NewTestRun(t, otf.TestRunCreateOptions{Status: otf.RunCanceled})

	ch := make(chan otf.Event, 1)

	spooler := &SpoolerDaemon{
		cancelations: make(chan Cancelation, 1),
		Watcher:      &testWatcher{ch: ch},
		Logger:       logr.Discard(),
	}

	go spooler.Start(context.Background())

	// send and receive cancelation
	ch <- otf.Event{Type: otf.EventRunCancel, Payload: want}
	got := <-spooler.GetCancelation()
	assert.Equal(t, Cancelation{Run: want, Forceful: false}, got)

	// send and receive forceful cancelation
	ch <- otf.Event{Type: otf.EventRunForceCancel, Payload: want}
	got = <-spooler.GetCancelation()
	assert.Equal(t, Cancelation{Run: want, Forceful: true}, got)
}
