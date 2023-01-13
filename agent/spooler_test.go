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
		&fakeSpoolerApp{runs: db, events: events},
		logr.Discard(),
		Config{},
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

func TestSpooler_handleEvent(t *testing.T) {
	tests := []struct {
		name                 string
		event                otf.Event
		config               Config
		wantRun              bool
		wantCancelation      bool
		wantForceCancelation bool
	}{
		{
			name: "handle run",
			event: otf.Event{
				Payload: otf.NewTestRun(t, otf.TestRunCreateOptions{
					Status: otf.RunPlanQueued,
				}),
			},
			wantRun: true,
		},
		{
			name: "internal agents skip agent-mode runs",
			event: otf.Event{
				Payload: otf.NewTestRun(t, otf.TestRunCreateOptions{
					ExecutionMode: otf.ExecutionModePtr(otf.AgentExecutionMode),
					Status:        otf.RunPlanQueued,
				}),
			},
			wantRun: false,
		},
		{
			name:   "external agents handle agent-mode runs",
			config: Config{External: true},
			event: otf.Event{
				Payload: otf.NewTestRun(t, otf.TestRunCreateOptions{
					ExecutionMode: otf.ExecutionModePtr(otf.AgentExecutionMode),
					Status:        otf.RunPlanQueued,
				}),
			},
			wantRun: true,
		},
		{
			name: "ignore runs not in queued state",
			event: otf.Event{
				Payload: otf.NewTestRun(t, otf.TestRunCreateOptions{
					Status: otf.RunPlanned,
				}),
			},
			wantRun: false,
		},
		{
			name: "handle cancelation",
			event: otf.Event{
				Type: otf.EventRunCancel,
				Payload: otf.NewTestRun(t, otf.TestRunCreateOptions{
					Status: otf.RunPlanning,
				}),
			},
			wantCancelation: true,
		},
		{
			name: "handle forceful cancelation",
			event: otf.Event{
				Type: otf.EventRunForceCancel,
				Payload: otf.NewTestRun(t, otf.TestRunCreateOptions{
					Status: otf.RunPlanning,
				}),
			},
			wantForceCancelation: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spooler := NewSpooler(nil, logr.Discard(), tt.config)
			spooler.handleEvent(tt.event)

			if tt.wantRun {
				assert.NotNil(t, <-spooler.GetRun())
			} else if tt.wantCancelation {
				assert.NotNil(t, <-spooler.GetCancelation())
			} else if tt.wantForceCancelation {
				if assert.Equal(t, 1, len(spooler.cancelations)) {
					got := <-spooler.GetCancelation()
					assert.True(t, got.Forceful)
				}
			} else {
				assert.Equal(t, 0, len(spooler.queue))
				assert.Equal(t, 0, len(spooler.cancelations))
			}
		})
	}
}
