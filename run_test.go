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
		wantPlanStatus  PlanStatus
		wantApplyStatus ApplyStatus
	}{
		{
			name:           "plan error",
			fromStatus:     RunPlanning,
			toStatus:       RunErrored,
			wantPlanStatus: PlanErrored,
		},
		{
			name:           "plan canceled",
			fromStatus:     RunPlanning,
			toStatus:       RunCanceled,
			wantPlanStatus: PlanCanceled,
		},
		{
			name:            "apply error",
			fromStatus:      RunApplying,
			toStatus:        RunErrored,
			wantApplyStatus: ApplyErrored,
		},
		{
			name:            "apply canceled",
			fromStatus:      RunApplying,
			toStatus:        RunCanceled,
			wantApplyStatus: ApplyCanceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Run{
				Status:           tt.fromStatus,
				StatusTimestamps: TimestampMap{},
				Plan: &Plan{
					StatusTimestamps: TimestampMap{},
				},
				Apply: &Apply{
					StatusTimestamps: TimestampMap{},
				},
			}

			r.UpdateStatus(tt.toStatus)

			assert.Equal(t, tt.wantPlanStatus, r.Plan.Status)
			assert.Equal(t, tt.wantApplyStatus, r.Apply.Status)
		})
	}
}

func TestRun_ForceCancelAvailableAt(t *testing.T) {
	run := &Run{
		StatusTimestamps: TimestampMap{},
		Plan: &Plan{
			StatusTimestamps: TimestampMap{},
		},
		Apply: &Apply{
			StatusTimestamps: TimestampMap{},
		},
	}

	run.UpdateStatus(RunCanceled)

	assert.NotZero(t, run.ForceCancelAvailableAt())
}

func TestRun_ForceCancelAvailableAt_IsZero(t *testing.T) {
	run := &Run{
		StatusTimestamps: TimestampMap{},
		Plan: &Plan{
			StatusTimestamps: TimestampMap{},
		},
		Apply: &Apply{
			StatusTimestamps: TimestampMap{},
		},
	}

	run.UpdateStatus(RunPending)

	assert.Zero(t, run.ForceCancelAvailableAt())
}
