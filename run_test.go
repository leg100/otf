package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_PlanThenApply(t *testing.T) {
	r := NewTestRun(t, "run-123", "ws-123", TestRunCreateOptions{})
	require.NoError(t, r.Enqueue())
	require.NoError(t, r.Start())

	// fake some planned changes
	reportService := &fakeReportService{ResourceReport{Additions: 1}}

	require.NoError(t, r.Finish(reportService, JobFinishOptions{}))
	require.NoError(t, r.ApplyRun())
	require.NoError(t, r.Start())
	require.NoError(t, r.Finish(reportService, JobFinishOptions{}))
}

func TestRun_PlanOnly(t *testing.T) {
	r := NewTestRun(t, "run-123", "ws-123", TestRunCreateOptions{
		Speculative: true,
	})
	require.NoError(t, r.Enqueue())
	require.NoError(t, r.Start())

	// fake some planned changes
	reportService := &fakeReportService{ResourceReport{Additions: 1}}

	require.NoError(t, r.Finish(reportService, JobFinishOptions{}))
	assert.Equal(t, RunPlannedAndFinished, r.Status())

	// ensure it cannot be applied
	require.Error(t, r.ApplyRun())
}

func TestRun_States(t *testing.T) {
	tests := []struct {
		status      RunStatus
		cancelable  bool
		discardable bool
		confirmable bool
		done        bool
	}{
		{
			status:     RunPending,
			cancelable: true,
		},
		{
			status: RunPlanQueued,
		},
		{
			status:     RunPlanning,
			cancelable: true,
		},
		{
			status:      RunPlanned,
			confirmable: true,
			discardable: true,
		},
		{
			status: RunPlannedAndFinished,
			done:   true,
		},
		{
			status: RunApplyQueued,
		},
		{
			status:     RunApplying,
			cancelable: true,
		},
		{
			status: RunApplied,
			done:   true,
		},
		{
			status: RunErrored,
			done:   true,
		},
		{
			status: RunDiscarded,
			done:   true,
		},
		{
			status: RunCanceled,
			done:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			run := NewTestRun(t, "run-123", "ws-123", TestRunCreateOptions{})
			run.setStateFromStatus(tt.status)

			assert.Equal(t, tt.cancelable, run.Cancelable())
			assert.Equal(t, tt.discardable, run.Discardable())
			assert.Equal(t, tt.confirmable, run.Confirmable())
			assert.Equal(t, tt.done, run.Done())
		})
	}
}

func TestRun_ForceCancelAvailableAt(t *testing.T) {
	run := NewTestRun(t, "run-123", "ws-123", TestRunCreateOptions{})
	run.setState(run.canceledState)

	assert.NotZero(t, run.ForceCancelAvailableAt())
}

func TestRun_ForceCancelAvailableAt_IsZero(t *testing.T) {
	run := NewTestRun(t, "run-123", "ws-123", TestRunCreateOptions{})

	assert.Zero(t, run.ForceCancelAvailableAt())
}
