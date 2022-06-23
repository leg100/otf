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
		wantPlanStatus  PhaseStatus
		wantApplyStatus PhaseStatus
	}{
		{
			name:            "plan error",
			fromStatus:      RunPlanning,
			toStatus:        RunErrored,
			wantPlanStatus:  PhaseErrored,
			wantApplyStatus: PhaseUnreachable,
		},
		{
			name:            "plan canceled",
			fromStatus:      RunPlanning,
			toStatus:        RunCanceled,
			wantPlanStatus:  PhaseCanceled,
			wantApplyStatus: PhaseUnreachable,
		},
		{
			name:            "apply error",
			fromStatus:      RunApplying,
			toStatus:        RunErrored,
			wantApplyStatus: PhaseErrored,
			wantPlanStatus:  PhasePending,
		},
		{
			name:            "apply canceled",
			fromStatus:      RunApplying,
			toStatus:        RunCanceled,
			wantApplyStatus: PhaseCanceled,
			wantPlanStatus:  PhasePending,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Run{
				status: tt.fromStatus,
				apply:  &Apply{},
			}
			r.plan = newPlan(r)
			r.apply = newApply(r)

			r.updateStatus(tt.toStatus)

			assert.Equal(t, tt.wantPlanStatus, r.plan.status)
			assert.Equal(t, tt.wantApplyStatus, r.apply.status)
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
		plan:  &Plan{},
		apply: &Apply{},
	}

	assert.NotZero(t, run.ForceCancelAvailableAt())
}

func TestRun_ForceCancelAvailableAt_IsZero(t *testing.T) {
	run := &Run{
		status: RunPending,
		plan:   &Plan{},
		apply:  &Apply{},
	}

	assert.Zero(t, run.ForceCancelAvailableAt())
}
