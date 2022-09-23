package agent

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestSpooler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// run[1-2] are in the DB; run[3-5] are events
	run1 := otf.NewTestRun(t, otf.TestRunCreateOptions{Status: otf.RunPlanQueued})
	run2 := otf.NewTestRun(t, otf.TestRunCreateOptions{Status: otf.RunPlanQueued})
	run3 := otf.NewTestRun(t, otf.TestRunCreateOptions{Status: otf.RunPlanQueued})
	run4 := otf.NewTestRun(t, otf.TestRunCreateOptions{Status: otf.RunCanceled})
	run5 := otf.NewTestRun(t, otf.TestRunCreateOptions{Status: otf.RunForceCanceled})
	db := []*otf.Run{run1, run2}
	events := make(chan otf.Event, 3)
	events <- otf.Event{Payload: run3}
	events <- otf.Event{Type: otf.EventRunCancel, Payload: run4}
	events <- otf.Event{Type: otf.EventRunForceCancel, Payload: run5}

	spooler := NewSpooler(
		ctx,
		&fakeSpoolerApp{runs: db, events: events},
		logr.Discard(),
		NewAgentOptions{Mode: InternalAgentMode},
	)
	errch := make(chan error)
	go func() { errch <- spooler.reinitialize(ctx) }()

	// expect to receive runs from DB in reverse order
	assert.Equal(t, run2, <-spooler.GetRun())
	assert.Equal(t, run1, <-spooler.GetRun())

	// expect afterwards to receive runs from events
	assert.Equal(t, run3, <-spooler.GetRun())
	assert.Equal(t, Cancelation{Run: run4, Forceful: false}, <-spooler.GetCancelation())
	assert.Equal(t, Cancelation{Run: run5, Forceful: true}, <-spooler.GetCancelation())
	cancel()
	assert.NoError(t, <-errch)
}
