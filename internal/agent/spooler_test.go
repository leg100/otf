package agent

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
)

func TestSpooler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// run[1-2] are in the DB; run[3-5] are events
	run1 := &run.Run{ExecutionMode: workspace.RemoteExecutionMode, Status: internal.RunPlanQueued}
	run2 := &run.Run{ExecutionMode: workspace.RemoteExecutionMode, Status: internal.RunPlanQueued}
	run3 := &run.Run{ExecutionMode: workspace.RemoteExecutionMode, Status: internal.RunPlanQueued}
	run4 := &run.Run{ExecutionMode: workspace.RemoteExecutionMode, Status: internal.RunCanceled}
	run5 := &run.Run{ExecutionMode: workspace.RemoteExecutionMode, Status: internal.RunForceCanceled}
	db := []*run.Run{run1, run2}
	events := make(chan pubsub.Event, 3)
	events <- pubsub.Event{Payload: run3}
	events <- pubsub.Event{Payload: run4}
	events <- pubsub.Event{Payload: run5}

	spooler := newSpooler(
		&fakeSpoolerApp{runs: db, events: events},
		logr.Discard(),
		Config{},
	)
	errch := make(chan error)
	go func() { errch <- spooler.reinitialize(ctx) }()

	// expect to receive runs from DB in reverse order
	assert.Equal(t, run2, <-spooler.getRun())
	assert.Equal(t, run1, <-spooler.getRun())

	// expect afterwards to receive runs from events
	assert.Equal(t, run3, <-spooler.getRun())
	assert.Equal(t, cancelation{Run: run4, Forceful: false}, <-spooler.getCancelation())
	assert.Equal(t, cancelation{Run: run5, Forceful: true}, <-spooler.getCancelation())
	cancel()
	assert.NoError(t, <-errch)
}

func TestSpooler_handleEvent(t *testing.T) {
	tests := []struct {
		name                 string
		event                pubsub.Event
		config               Config
		wantRun              bool
		wantCancelation      bool
		wantForceCancelation bool
	}{
		{
			name: "handle run",
			event: pubsub.Event{
				Payload: &run.Run{
					ExecutionMode: workspace.RemoteExecutionMode,
					Status:        internal.RunPlanQueued,
				},
			},
			wantRun: true,
		},
		{
			name:   "internal agents skip agent-mode runs",
			config: Config{External: false},
			event: pubsub.Event{
				Payload: &run.Run{
					ExecutionMode: workspace.AgentExecutionMode,
				},
			},
			wantRun: false,
		},
		{
			name:   "external agents handle agent-mode runs",
			config: Config{External: true},
			event: pubsub.Event{
				Payload: &run.Run{
					ExecutionMode: workspace.AgentExecutionMode,
					Status:        internal.RunPlanQueued,
				},
			},
			wantRun: true,
		},
		{
			name: "ignore runs not in queued state",
			event: pubsub.Event{
				Payload: &run.Run{
					ExecutionMode: workspace.RemoteExecutionMode,
					Status:        internal.RunPlanned,
				},
			},
			wantRun: false,
		},
		{
			name: "handle cancelation",
			event: pubsub.Event{
				Payload: &run.Run{
					ExecutionMode: workspace.RemoteExecutionMode,
					Status:        internal.RunCanceled,
				},
			},
			wantCancelation: true,
		},
		{
			name: "handle forceful cancelation",
			event: pubsub.Event{
				Payload: &run.Run{
					ExecutionMode: workspace.RemoteExecutionMode,
					Status:        internal.RunForceCanceled,
				},
			},
			wantForceCancelation: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spooler := newSpooler(nil, logr.Discard(), tt.config)
			spooler.handleEvent(tt.event)

			if tt.wantRun {
				assert.NotNil(t, <-spooler.getRun())
			} else if tt.wantCancelation {
				assert.NotNil(t, <-spooler.getCancelation())
			} else if tt.wantForceCancelation {
				if assert.Equal(t, 1, len(spooler.cancelations)) {
					got := <-spooler.getCancelation()
					assert.True(t, got.Forceful)
				}
			} else {
				assert.Equal(t, 0, len(spooler.queue))
				assert.Equal(t, 0, len(spooler.cancelations))
			}
		})
	}
}
