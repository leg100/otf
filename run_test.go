package ots

import (
	"testing"

	tfe "github.com/leg100/go-tfe"
	"github.com/stretchr/testify/assert"
)

// TestRun_UpdateStatus tests that UpdateStatus correctly updates the status of
// the run's plan and apply (there is little point to testing the status of the
// run itself because there is no conditional logic to this assignment).
func TestRun_UpdateStatus(t *testing.T) {
	tests := []struct {
		name            string
		fromStatus      tfe.RunStatus
		toStatus        tfe.RunStatus
		wantPlanStatus  tfe.PlanStatus
		wantApplyStatus tfe.ApplyStatus
	}{
		{
			name:           "plan error",
			fromStatus:     tfe.RunPlanning,
			toStatus:       tfe.RunErrored,
			wantPlanStatus: tfe.PlanErrored,
		},
		{
			name:           "plan canceled",
			fromStatus:     tfe.RunPlanning,
			toStatus:       tfe.RunCanceled,
			wantPlanStatus: tfe.PlanCanceled,
		},
		{
			name:            "apply error",
			fromStatus:      tfe.RunApplying,
			toStatus:        tfe.RunErrored,
			wantApplyStatus: tfe.ApplyErrored,
		},
		{
			name:            "apply canceled",
			fromStatus:      tfe.RunApplying,
			toStatus:        tfe.RunCanceled,
			wantApplyStatus: tfe.ApplyCanceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Run{
				Status:           tt.fromStatus,
				StatusTimestamps: &tfe.RunStatusTimestamps{},
				Plan: &Plan{
					StatusTimestamps: &tfe.PlanStatusTimestamps{},
				},
				Apply: &Apply{
					StatusTimestamps: &tfe.ApplyStatusTimestamps{},
				},
			}

			r.UpdateStatus(tt.toStatus)

			assert.Equal(t, tt.wantPlanStatus, r.Plan.Status)
			assert.Equal(t, tt.wantApplyStatus, r.Apply.Status)
		})
	}
}
