package run

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	tests := []struct {
		name            string
		run             *Run
		planningTimeout time.Duration
		applyingTimeout time.Duration
		// whether run has been timed out
		timedout bool
	}{
		{
			name: "planning timeout exceeded",
			run: &Run{
				Status: RunPlanning,
				StatusTimestamps: []StatusTimestamp{
					{
						Status:    RunPlanning,
						Timestamp: time.Now().Add(-2 * time.Hour),
					},
				},
			},
			planningTimeout: time.Hour,
			timedout:        true,
		},
		{
			name: "applying timeout exceeded",
			run: &Run{
				Status: RunApplying,
				StatusTimestamps: []StatusTimestamp{
					{
						Status:    RunApplying,
						Timestamp: time.Now().Add(-2 * time.Hour),
					},
				},
			},
			applyingTimeout: time.Hour,
			timedout:        true,
		},
		{
			name: "planning timeout not exceeded",
			run: &Run{
				Status: RunPlanning,
				StatusTimestamps: []StatusTimestamp{
					{
						Status:    RunPlanning,
						Timestamp: time.Now().Add(time.Hour),
					},
				},
			},
			planningTimeout: 2 * time.Hour,
			timedout:        false,
		},
		{
			name: "applying timeout not exceeded",
			run: &Run{
				Status: RunApplying,
				StatusTimestamps: []StatusTimestamp{
					{
						Status:    RunApplying,
						Timestamp: time.Now().Add(time.Hour),
					},
				},
			},
			applyingTimeout: 2 * time.Hour,
			timedout:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeTimeoutRunClient{run: tt.run}
			timeout := &Timeout{
				Runs:            client,
				PlanningTimeout: tt.planningTimeout,
			}
			timeout.check(context.Background())
			assert.Equal(t, client.timeout, tt.timedout)
		})
	}
}

type fakeTimeoutRunClient struct {
	run     *Run
	timeout bool
}

func (f *fakeTimeoutRunClient) List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	return resource.NewPage([]*Run{f.run}, resource.PageOptions{}, nil), nil
}

func (f *fakeTimeoutRunClient) FinishPhase(ctx context.Context, runID string, phase internal.PhaseType, opts PhaseFinishOptions) (*Run, error) {
	f.timeout = true
	return nil, nil
}
