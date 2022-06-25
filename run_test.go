package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRun_Cancel_Pending(t *testing.T) {
	run := NewTestRun(t, "run-123", "ws-123", TestRunCreateOptions{Status: RunPending})
	enqueue, err := run.Cancel()
	require.NoError(t, err)
	assert.False(t, enqueue)
	assert.NotZero(t, run.forceCancelAvailableAt)
}

func TestRun_Cancel_Planning(t *testing.T) {
	run := NewTestRun(t, "run-123", "ws-123", TestRunCreateOptions{Status: RunPlanning})
	enqueue, err := run.Cancel()
	require.NoError(t, err)
	assert.True(t, enqueue)
	assert.NotZero(t, run.forceCancelAvailableAt)
}
