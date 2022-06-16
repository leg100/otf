package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRun_UpdateStatus tests that UpdateStatus correctly updates the status of
// the run's plan and apply (there is little point to testing the status of the
// run itself because there is no conditional logic to this assignment).
func TestRun_UpdateStatus(t *testing.T) {
	tests := []struct {
		name            string
		fromStatus      RunStatus
		toStatus        RunStatus
		wantPlanStatus  JobStatus
		wantApplyStatus JobStatus
	}{
		{
			name:            "plan error",
			fromStatus:      RunPlanning,
			toStatus:        RunErrored,
			wantPlanStatus:  JobErrored,
			wantApplyStatus: JobUnreachable,
		},
		{
			name:            "plan canceled",
			fromStatus:      RunPlanning,
			toStatus:        RunCanceled,
			wantPlanStatus:  JobCanceled,
			wantApplyStatus: JobUnreachable,
		},
		{
			name:            "apply error",
			fromStatus:      RunApplying,
			toStatus:        RunErrored,
			wantApplyStatus: JobErrored,
			wantPlanStatus:  JobPending,
		},
		{
			name:            "apply canceled",
			fromStatus:      RunApplying,
			toStatus:        RunCanceled,
			wantApplyStatus: JobCanceled,
			wantPlanStatus:  JobPending,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Run{
				status: tt.fromStatus,
				Apply:  &Apply{},
			}
			r.Plan = newPlan(r)
			r.Apply = newApply(r)

			r.updateStatus(tt.toStatus)

			assert.Equal(t, tt.wantPlanStatus, r.Plan.status)
			assert.Equal(t, tt.wantApplyStatus, r.Apply.status)
		})
	}
}

func TestRun_ForceCancelAvailableAt(t *testing.T) {
	run := &Run{
		status: RunCanceled,
		statusTimestamps: []RunStatusTimestamp{
			{
				Status:    RunCanceled,
				Timestamp: CurrentTimestamp(),
			},
		},
		Plan:  &Plan{},
		Apply: &Apply{},
	}

	assert.NotZero(t, run.ForceCancelAvailableAt())
}

func TestRun_ForceCancelAvailableAt_IsZero(t *testing.T) {
	run := &Run{
		status: RunPending,
		Plan:   &Plan{},
		Apply:  &Apply{},
	}

	assert.Zero(t, run.ForceCancelAvailableAt())
}
