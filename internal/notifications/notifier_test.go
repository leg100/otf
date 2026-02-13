package notifications

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifier_handleRun(t *testing.T) {
	ctx := context.Background()
	ws1 := testutils.ParseID(t, "ws-matching")
	ws2 := testutils.ParseID(t, "ws-zzz")

	queuedRunEvent := pubsub.Event[*run.Event]{Payload: &run.Event{
		Status:      runstatus.PlanQueued,
		WorkspaceID: ws1,
	}}
	planningRunID := resource.NewTfeID(resource.RunKind)
	planningRunEvent := pubsub.Event[*run.Event]{Payload: &run.Event{
		ID:          planningRunID,
		Status:      runstatus.Planning,
		WorkspaceID: ws1,
	}}
	disabledConfig := &Config{
		URL:         new(""),
		WorkspaceID: ws1,
		Triggers:    []Trigger{TriggerPlanning},
	}
	enabledConfig := &Config{
		URL:         new(""),
		WorkspaceID: ws1,
		Enabled:     true,
		Triggers:    []Trigger{TriggerPlanning},
	}
	configWithNoTriggers := &Config{
		URL:         new(""),
		Enabled:     true,
		WorkspaceID: ws1,
	}
	configWithDifferentTriggers := &Config{
		URL:         new(""),
		Enabled:     true,
		WorkspaceID: ws1,
		Triggers:    []Trigger{TriggerApplying},
	}
	configForDifferentWorkspace := &Config{
		URL:         new(""),
		WorkspaceID: ws2,
		Enabled:     true,
		Triggers:    []Trigger{TriggerPlanning},
	}

	tests := []struct {
		name  string
		event pubsub.Event[*run.Event]
		cfg   *Config
		want  *notification
	}{
		{"ignore queued run", queuedRunEvent, enabledConfig, nil},
		{"no matching configs", planningRunEvent, configForDifferentWorkspace, nil},
		{"disabled config", planningRunEvent, disabledConfig, nil},
		{"enabled but no triggers", planningRunEvent, configWithNoTriggers, nil},
		{"enabled but mis-matching triggers", planningRunEvent, configWithDifferentTriggers, nil},
		{"matching trigger", planningRunEvent, enabledConfig, &notification{
			event:   planningRunEvent,
			trigger: TriggerPlanning,
			config:  enabledConfig,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			published := make(chan *notification, 100)
			notifier := &Notifier{
				Logger:     logr.Discard(),
				workspaces: &fakeWorkspaceService{},
				runs:       &fakeRunService{},
				system:     &fakeHostnameService{},
				cache:      newTestCache(t, &fakeFactory{published}, tt.cfg),
			}

			err := notifier.handleRunEvent(ctx, tt.event)
			require.NoError(t, err)
			if tt.want != nil {
				assert.Equal(t, tt.want, <-published)
			} else {
				assert.Equal(t, 0, len(published))
			}
		})
	}
}

// TestNotifier_handleRun_multiple tests handleRun() publishing multiple
// notifications
func TestNotifier_handleRun_multiple(t *testing.T) {
	ctx := context.Background()
	ws1 := testutils.ParseID(t, "ws-123")

	planningRunEvent := pubsub.Event[*run.Event]{Payload: &run.Event{
		ID:          resource.NewTfeID(resource.RunKind),
		Status:      runstatus.Planning,
		WorkspaceID: ws1,
	}}
	config1 := newTestConfig(t, ws1, DestinationGCPPubSub, "", TriggerPlanning)
	config2 := newTestConfig(t, ws1, DestinationSlack, "", TriggerPlanning)

	published := make(chan *notification, 2)
	notifier := &Notifier{
		Logger:     logr.Discard(),
		workspaces: &fakeWorkspaceService{},
		runs:       &fakeRunService{},
		system:     &fakeHostnameService{},
		cache:      newTestCache(t, &fakeFactory{published}, config1, config2),
	}

	err := notifier.handleRunEvent(ctx, planningRunEvent)
	require.NoError(t, err)

	// Expect two notifications to be published
	var got []*notification
	got = append(got, <-published)
	got = append(got, <-published)

	// One notification for gcp pub sub
	want := &notification{
		event:   planningRunEvent,
		config:  config1,
		trigger: TriggerPlanning,
	}
	assert.Contains(t, got, want)

	// One notification for slack
	want = &notification{
		event:   planningRunEvent,
		config:  config2,
		trigger: TriggerPlanning,
	}
	assert.Contains(t, got, want)
}

func TestNotifier_handleConfig(t *testing.T) {
	ws1 := testutils.ParseID(t, "ws-123")

	notifier := &Notifier{
		Logger:     logr.Discard(),
		workspaces: &fakeWorkspaceService{},
		system:     &fakeHostnameService{},
		cache:      newTestCache(t, &fakeFactory{}),
	}

	// Add config, should result in cache size of 1
	config1 := newTestConfig(t, ws1, DestinationGCPPubSub, "gcppubsub://proj1/topic1", TriggerPlanning)
	err := notifier.handleConfigEvent(pubsub.Event[*Config]{Payload: config1, Type: pubsub.CreatedEvent})
	require.NoError(t, err)
	assert.Len(t, notifier.cache.configs, 1)
	assert.Len(t, notifier.cache.clients, 1)

	// Update config url, cache size should still be 1
	updated := newTestConfig(t, ws1, DestinationGCPPubSub, "gcppubsub://proj2/topic2", TriggerPlanning)
	updated.ID = config1.ID
	err = notifier.handleConfigEvent(pubsub.Event[*Config]{Payload: updated, Type: pubsub.UpdatedEvent})
	require.NoError(t, err)
	assert.Len(t, notifier.cache.configs, 1)
	assert.Len(t, notifier.cache.clients, 1)

	// Remove config, cache size should be 0
	err = notifier.handleConfigEvent(pubsub.Event[*Config]{Payload: updated, Type: pubsub.DeletedEvent})
	require.NoError(t, err)
	assert.Len(t, notifier.cache.configs, 0)
	assert.Len(t, notifier.cache.clients, 0)
}
