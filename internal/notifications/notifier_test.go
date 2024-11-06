package notifications

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifier_handleRun(t *testing.T) {
	ctx := context.Background()
	ws1 := resource.ParseID("ws-matching")
	ws2 := resource.ParseID("ws-zzz")

	queuedRun := &run.Run{
		Status:      run.RunPlanQueued,
		WorkspaceID: ws1,
	}
	planningRun := &run.Run{
		Status:      run.RunPlanning,
		WorkspaceID: ws1,
	}
	disabledConfig := &Config{
		URL:         internal.String(""),
		WorkspaceID: ws1,
		Triggers:    []Trigger{TriggerPlanning},
	}
	enabledConfig := &Config{
		URL:         internal.String(""),
		WorkspaceID: ws1,
		Enabled:     true,
		Triggers:    []Trigger{TriggerPlanning},
	}
	configWithNoTriggers := &Config{
		URL:         internal.String(""),
		Enabled:     true,
		WorkspaceID: ws1,
	}
	configWithDifferentTriggers := &Config{
		URL:         internal.String(""),
		Enabled:     true,
		WorkspaceID: ws1,
		Triggers:    []Trigger{TriggerApplying},
	}
	configForDifferentWorkspace := &Config{
		URL:         internal.String(""),
		WorkspaceID: ws2,
		Enabled:     true,
		Triggers:    []Trigger{TriggerPlanning},
	}

	tests := []struct {
		name          string
		run           *run.Run
		cfg           *Config
		wantPublished bool
	}{
		{"ignore queued run", queuedRun, enabledConfig, false},
		{"no matching configs", planningRun, configForDifferentWorkspace, false},
		{"disabled config", planningRun, disabledConfig, false},
		{"enabled but no triggers", planningRun, configWithNoTriggers, false},
		{"enabled but mis-matching triggers", planningRun, configWithDifferentTriggers, false},
		{"matching trigger", planningRun, enabledConfig, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			published := make(chan *run.Run, 100)
			notifier := &Notifier{
				Logger:     logr.Discard(),
				workspaces: &fakeWorkspaceService{},
				system:     &fakeHostnameService{},
				cache:      newTestCache(t, &fakeFactory{published}, tt.cfg),
			}

			err := notifier.handleRun(ctx, tt.run)
			require.NoError(t, err)
			if tt.wantPublished {
				assert.Equal(t, tt.run, <-published)
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
	ws1 := resource.ParseID("ws-123")

	planningRun := &run.Run{
		Status:      run.RunPlanning,
		WorkspaceID: ws1,
	}
	config1 := newTestConfig(t, ws1, DestinationGCPPubSub, "", TriggerPlanning)
	config2 := newTestConfig(t, ws1, DestinationSlack, "", TriggerPlanning)

	published := make(chan *run.Run, 2)
	notifier := &Notifier{
		Logger:     logr.Discard(),
		workspaces: &fakeWorkspaceService{},
		system:     &fakeHostnameService{},
		cache:      newTestCache(t, &fakeFactory{published}, config1, config2),
	}

	err := notifier.handleRun(ctx, planningRun)
	require.NoError(t, err)
	assert.Equal(t, planningRun, <-published)
	assert.Equal(t, planningRun, <-published)
}

func TestNotifier_handleConfig(t *testing.T) {
	ctx := context.Background()
	ws1 := resource.ParseID("ws-123")

	notifier := &Notifier{
		Logger:     logr.Discard(),
		workspaces: &fakeWorkspaceService{},
		system:     &fakeHostnameService{},
		cache:      newTestCache(t, &fakeFactory{}),
	}

	// Add config, should result in cache size of 1
	config1 := newTestConfig(t, ws1, DestinationGCPPubSub, "gcppubsub://proj1/topic1", TriggerPlanning)
	err := notifier.handleConfig(ctx, pubsub.Event[*Config]{Payload: config1, Type: pubsub.CreatedEvent})
	require.NoError(t, err)
	assert.Len(t, notifier.cache.configs, 1)
	assert.Len(t, notifier.cache.clients, 1)

	// Update config url, cache size should still be 1
	updated := newTestConfig(t, ws1, DestinationGCPPubSub, "gcppubsub://proj2/topic2", TriggerPlanning)
	updated.ID = config1.ID
	err = notifier.handleConfig(ctx, pubsub.Event[*Config]{Payload: updated, Type: pubsub.UpdatedEvent})
	require.NoError(t, err)
	assert.Len(t, notifier.cache.configs, 1)
	assert.Len(t, notifier.cache.clients, 1)

	// Remove config, cache size should be 0
	err = notifier.handleConfig(ctx, pubsub.Event[*Config]{Payload: updated, Type: pubsub.DeletedEvent})
	require.NoError(t, err)
	assert.Len(t, notifier.cache.configs, 0)
	assert.Len(t, notifier.cache.clients, 0)
}
