package agent

import (
	"context"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
)

// The mapper maps the state of runs to jobs and vice-versa, and creates jobs when
// runs enter certain states.
//
// Only one mapper should be active on an OTF cluster.
type mapper struct {
	// RunService for retrieving list of runs at startup
	run.RunService
	// Subscriber for receiving stream of run events
	pubsub.Subscriber
}

// Start the mapper. Should be invoked in a go routine.
func (m *mapper) Start(ctx context.Context) error {
	// Subscribe to run events and unsubscribe before returning.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sub, err := m.Subscribe(ctx, "job-mapper-")
	if err != nil {
		return err
	}
	for event := range sub {
		run, ok := event.Payload.(*run.Run)
		if !ok {
			// Skip non-run events
			continue
		}
		if event.Type == pubsub.DeletedEvent {
			// Skip deleted run events
			continue
		}
		if err := m.handleRun(ctx, run); err != nil {
			return err
		}
	}
	return pubsub.ErrSubscriptionTerminated
}

func (m *mapper) handleRun(ctx context.Context, run *run.Run) error {
	return nil
}
