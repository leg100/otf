package trigger

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
)

// triggerer creates runs when runs on a sourceable workspace complete.
type triggerer struct {
	client client
	logger logr.Logger
}

type client interface {
	ListRunTriggers(ctx context.Context, opts ListOptions) ([]*trigger, error)
	WatchRuns(context.Context) (<-chan pubsub.Event[*run.Event], func(), error)
	CreateRun(context.Context, resource.TfeID, run.CreateOptions) (*run.Run, error)
}

func (t *triggerer) Start(ctx context.Context) error {
	// Subscribe to run events
	sub, unsub, err := t.client.WatchRuns(ctx)
	if err != nil {
		return fmt.Errorf("watching runs: %w", err)
	}
	defer unsub()

	for runEvent := range sub {
		if err := t.process(ctx, runEvent); err != nil {
			return err
		}
	}

	return nil
}

func (t *triggerer) process(ctx context.Context, runEvent pubsub.Event[*run.Event]) error {
	// Only interested in run updates
	if runEvent.Type != pubsub.UpdatedEvent {
		return nil
	}
	// Only interested in finished runs
	if !runstatus.Done(runEvent.Payload.Status) {
		return nil
	}
	// Only interested in runs that have been successfully applied
	if runEvent.Payload.Status != runstatus.Applied {
		return nil
	}

	// Look up workspaces triggered by the finished run's workspace.
	triggers, err := t.client.ListRunTriggers(ctx, ListOptions{
		WorkspaceID: runEvent.Payload.WorkspaceID,
		Direction:   Outbound,
	})
	if err != nil {
		return fmt.Errorf("listing triggers for finished run: %w", err)
	}

	for _, trigger := range triggers {
		t.logger.Info("triggering run in connected workspace")

		_, err := t.client.CreateRun(ctx, trigger.WorkspaceID, run.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
