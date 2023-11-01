package agent

import (
	"context"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
)

// Allocator allocates jobs to agents. Only one allocator runs across an entire
// OTF deployment. The allocator at startup retrieves a list of all incomplete
// runs and ensures there are jobs corresponding to them. If not, jobs are
// created. The allocator then listens for run events, creating jobs as
// necessary.
type Allocator struct {
	// RunService for retrieving list of runs at startup
	run.RunService
	// Subscriber for receiving stream of run events
	pubsub.Subscriber
}

// Start the allocator. Should be invoked in a go routine.
func (a *Allocator) Start(ctx context.Context) error {
	// Subscribe to run events and unsubscribe before returning.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sub, err := a.Subscribe(ctx, "allocator-")
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
		if err := a.handleRun(ctx, run); err != nil {
			return err
		}
	}
	return pubsub.ErrSubscriptionTerminated
}

func (a *Allocator) handleRun(ctx context.Context, run *run.Run) error {
	return nil
}
