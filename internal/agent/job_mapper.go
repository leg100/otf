package agent

import (
	"context"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
)

// The jobMapper is responsible for mapping jobs to runs. When a run
// enters certain states, the jobMapper creates corresponding jobs, e.g.
// when the run enters the plan_queued state, the jobMapper creates a
// plan job. Inversely, when a job enters certain states, the jobMapper
// updates the corresponding run state, e.g. when a plan job enters the running
// state, the jobMapper updates the run state to planning.
//
// Only one jobMapper should be active on an OTF cluster.
type jobMapper struct {
	// RunService for retrieving list of runs at startup
	run.RunService
	// Subscriber for receiving stream of run events
	pubsub.Subscriber
}

// Start the mapper. Should be invoked in a go routine.
func (m *jobMapper) Start(ctx context.Context) error {
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

func (m *jobMapper) handleRun(ctx context.Context, run *run.Run) error {
	return nil
}
