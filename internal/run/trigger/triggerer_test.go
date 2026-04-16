package trigger

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTriggerer_process(t *testing.T) {
	ws1 := resource.NewTfeID(resource.WorkspaceKind)
	ws2 := resource.NewTfeID(resource.WorkspaceKind)

	tests := []struct {
		name      string
		event     pubsub.Event[*run.Event]
		triggers  []*Trigger
		wantRuns  int
		wantError bool
	}{
		{
			name: "ignores created events",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.CreatedEvent,
				Payload: &run.Event{Status: runstatus.Applied, WorkspaceID: ws1},
			},
			wantRuns: 0,
		},
		{
			name: "ignores deleted events",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.DeletedEvent,
				Payload: &run.Event{Status: runstatus.Applied, WorkspaceID: ws1},
			},
			wantRuns: 0,
		},
		{
			name: "ignores non-done statuses",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.UpdatedEvent,
				Payload: &run.Event{Status: runstatus.Planning, WorkspaceID: ws1},
			},
			wantRuns: 0,
		},
		{
			name: "ignores errored runs",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.UpdatedEvent,
				Payload: &run.Event{Status: runstatus.Errored, WorkspaceID: ws1},
			},
			wantRuns: 0,
		},
		{
			name: "ignores canceled runs",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.UpdatedEvent,
				Payload: &run.Event{Status: runstatus.Canceled, WorkspaceID: ws1},
			},
			wantRuns: 0,
		},
		{
			name: "ignores discarded runs",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.UpdatedEvent,
				Payload: &run.Event{Status: runstatus.Discarded, WorkspaceID: ws1},
			},
			wantRuns: 0,
		},
		{
			name: "ignores planned_and_finished runs",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.UpdatedEvent,
				Payload: &run.Event{Status: runstatus.PlannedAndFinished, WorkspaceID: ws1},
			},
			wantRuns: 0,
		},
		{
			name: "no triggers for workspace",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.UpdatedEvent,
				Payload: &run.Event{Status: runstatus.Applied, WorkspaceID: ws1},
			},
			triggers: []*Trigger{},
			wantRuns: 0,
		},
		{
			name: "creates run in triggered workspace",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.UpdatedEvent,
				Payload: &run.Event{Status: runstatus.Applied, WorkspaceID: ws1},
			},
			triggers: []*Trigger{
				{WorkspaceID: ws2, SourceableWorkspaceID: ws1},
			},
			wantRuns: 1,
		},
		{
			name: "creates runs in multiple triggered workspaces",
			event: pubsub.Event[*run.Event]{
				Type:    pubsub.UpdatedEvent,
				Payload: &run.Event{Status: runstatus.Applied, WorkspaceID: ws1},
			},
			triggers: []*Trigger{
				{WorkspaceID: ws2, SourceableWorkspaceID: ws1},
				{WorkspaceID: resource.NewTfeID(resource.WorkspaceKind), SourceableWorkspaceID: ws1},
			},
			wantRuns: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeTriggererClient{triggers: tt.triggers}
			triggerer := &Triggerer{
				Client: fake,
				Logger: logr.Discard(),
			}

			err := triggerer.process(t.Context(), tt.event)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantRuns, len(fake.createdRuns))
		})
	}
}

func TestTriggerer_Start(t *testing.T) {
	ws1 := resource.NewTfeID(resource.WorkspaceKind)
	ws2 := resource.NewTfeID(resource.WorkspaceKind)

	events := make(chan pubsub.Event[*run.Event], 2)
	events <- pubsub.Event[*run.Event]{
		Type:    pubsub.UpdatedEvent,
		Payload: &run.Event{Status: runstatus.Applied, WorkspaceID: ws1},
	}
	events <- pubsub.Event[*run.Event]{
		Type:    pubsub.UpdatedEvent,
		Payload: &run.Event{Status: runstatus.Errored, WorkspaceID: ws1},
	}
	close(events)

	fake := &fakeTriggererClient{
		events: events,
		triggers: []*Trigger{
			{WorkspaceID: ws2, SourceableWorkspaceID: ws1},
		},
	}
	triggerer := &Triggerer{
		Client: fake,
		Logger: logr.Discard(),
	}

	err := triggerer.Start(t.Context())
	require.NoError(t, err)
	// Only the applied run should have triggered a new run.
	assert.Equal(t, 1, len(fake.createdRuns))
	assert.Equal(t, ws2, fake.createdRuns[0])
}

type fakeTriggererClient struct {
	events   <-chan pubsub.Event[*run.Event]
	triggers []*Trigger
	// createdRuns records workspace IDs for which CreateRun was called.
	createdRuns []resource.TfeID
}

func (f *fakeTriggererClient) WatchRuns(ctx context.Context) (<-chan pubsub.Event[*run.Event], func(), error) {
	return f.events, func() {}, nil
}

func (f *fakeTriggererClient) ListRunTriggers(ctx context.Context, opts ListOptions) ([]*Trigger, error) {
	return f.triggers, nil
}

func (f *fakeTriggererClient) CreateRun(ctx context.Context, workspaceID resource.TfeID, opts run.CreateOptions) (*run.Run, error) {
	f.createdRuns = append(f.createdRuns, workspaceID)
	return &run.Run{ID: resource.NewTfeID(resource.RunKind)}, nil
}
