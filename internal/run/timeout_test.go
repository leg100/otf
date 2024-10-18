package run

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	tests := []struct {
		name            string
		run             *Run
		planningTimeout time.Duration
		applyingTimeout time.Duration
		// expect run to be canceled or not
		canceled bool
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
			canceled:        true,
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
			canceled:        true,
		},
		{
			name: "planning timeout not exceeded",
			run: &Run{
				Status: RunPlanning,
				StatusTimestamps: []StatusTimestamp{
					{
						Status:    RunPlanning,
						Timestamp: time.Now().Add(-time.Hour),
					},
				},
			},
			planningTimeout: 2 * time.Hour,
			canceled:        false,
		},
		{
			name: "applying timeout not exceeded",
			run: &Run{
				Status: RunApplying,
				StatusTimestamps: []StatusTimestamp{
					{
						Status:    RunApplying,
						Timestamp: time.Now().Add(-time.Hour),
					},
				},
			},
			applyingTimeout: 2 * time.Hour,
			canceled:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeTimeoutRunClient{run: tt.run}
			timeout := &Timeout{
				Runs:            client,
				PlanningTimeout: tt.planningTimeout,
				ApplyingTimeout: tt.applyingTimeout,
			}
			timeout.check(context.Background())
			assert.Equal(t, tt.canceled, client.canceled)
		})
	}
}

type fakeTimeoutRunClient struct {
	run      *Run
	canceled bool
}

func (f *fakeTimeoutRunClient) List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	return resource.NewPage([]*Run{f.run}, resource.PageOptions{}, nil), nil
}

func (f *fakeTimeoutRunClient) Cancel(ctx context.Context, runID string) error {
	f.canceled = true
	return nil
}
