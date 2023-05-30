package notifications

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifier_handleRun(t *testing.T) {
	ctx := context.Background()

	queuedRun := &run.Run{
		Status:      internal.RunPlanQueued,
		WorkspaceID: "ws-matching",
	}
	planningRun := &run.Run{
		Status:      internal.RunPlanning,
		WorkspaceID: "ws-matching",
	}
	disabledConfig := &Config{
		URL:         internal.String(""),
		WorkspaceID: "ws-matching",
		Triggers:    []Trigger{TriggerPlanning},
	}
	enabledConfig := &Config{
		URL:         internal.String(""),
		WorkspaceID: "ws-matching",
		Enabled:     true,
		Triggers:    []Trigger{TriggerPlanning},
	}
	configWithNoTriggers := &Config{
		URL:         internal.String(""),
		Enabled:     true,
		WorkspaceID: "ws-matching",
	}
	configWithDifferentTriggers := &Config{
		URL:         internal.String(""),
		Enabled:     true,
		WorkspaceID: "ws-matching",
		Triggers:    []Trigger{TriggerApplying},
	}
	configForDifferentWorkspace := &Config{
		URL:         internal.String(""),
		WorkspaceID: "ws-zzz",
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
			notifier := newTestNotifier(t, &fakeFactory{published}, tt.cfg)

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
	planningRun := &run.Run{
		Status:      internal.RunPlanning,
		WorkspaceID: "ws-123",
	}
	config1 := newTestConfig(t, "ws-123", DestinationGCPPubSub, "", TriggerPlanning)
	config2 := newTestConfig(t, "ws-123", DestinationSlack, "", TriggerPlanning)

	published := make(chan *run.Run, 2)
	notifier := newTestNotifier(t, &fakeFactory{published}, config1, config2)

	err := notifier.handleRun(ctx, planningRun)
	require.NoError(t, err)
	assert.Equal(t, planningRun, <-published)
	assert.Equal(t, planningRun, <-published)
}
