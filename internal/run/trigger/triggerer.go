package trigger

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
)

// Triggerer creates runs when runs on a sourceable workspace complete.
type Triggerer struct {
	Client client
	Logger logr.Logger
}

type client interface {
	ListRunTriggers(ctx context.Context, opts ListOptions) ([]*Trigger, error)
	WatchRuns(context.Context) (<-chan pubsub.Event[*run.Event], func(), error)
	CreateRun(context.Context, resource.TfeID, run.CreateOptions) (*run.Run, error)
}

func (t *Triggerer) Start(ctx context.Context) error {
	// Subscribe to run events
	sub, unsub, err := t.Client.WatchRuns(ctx)
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

func (t *Triggerer) process(ctx context.Context, runEvent pubsub.Event[*run.Event]) error {
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
	triggers, err := t.Client.ListRunTriggers(ctx, ListOptions{
		WorkspaceID: runEvent.Payload.WorkspaceID,
		Direction:   Outbound,
	})
	if err != nil {
		return fmt.Errorf("listing triggers for finished run: %w", err)
	}

	for _, trigger := range triggers {
		t.Logger.Info(
			"triggering run in connected workspace",
			"trigger", trigger,
			"triggering_run_id", runEvent.Payload.ID,
		)

		_, err := t.Client.CreateRun(ctx, trigger.WorkspaceID, run.CreateOptions{
			CreatedBy:       runEvent.Payload.CreatedBy,
			Source:          source.Trigger,
			TriggeringRunID: &runEvent.Payload.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
