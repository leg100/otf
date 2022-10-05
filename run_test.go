package otf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_States(t *testing.T) {
	t.Run("pending", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{})

		require.Equal(t, RunPending, run.status)
		require.Equal(t, PhasePending, run.plan.status)
		require.Equal(t, PhasePending, run.apply.status)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{})

		require.NoError(t, run.EnqueuePlan(context.Background(), &FakeWorkspaceLockService{}))

		require.Equal(t, RunPlanQueued, run.status)
		require.Equal(t, PhaseQueued, run.plan.status)
		require.Equal(t, PhasePending, run.apply.status)
	})

	t.Run("start plan", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{Status: RunPlanQueued})

		require.NoError(t, run.Start(PlanPhase))

		require.Equal(t, RunPlanning, run.status)
		require.Equal(t, PhaseRunning, run.plan.status)
		require.Equal(t, PhasePending, run.apply.status)
	})

	t.Run("finish plan", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{Status: RunPlanning})

		require.NoError(t, run.Finish(PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunPlannedAndFinished, run.status)
		require.Equal(t, PhaseFinished, run.plan.status)
		require.Equal(t, PhaseUnreachable, run.apply.status)
	})

	t.Run("finish plan with errors", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{Status: RunPlanning})

		require.NoError(t, run.Finish(PlanPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, RunErrored, run.status)
		require.Equal(t, PhaseErrored, run.plan.status)
		require.Equal(t, PhaseUnreachable, run.apply.status)
	})

	t.Run("finish plan with changes", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{Status: RunPlanning})

		run.plan.ResourceReport = &ResourceReport{Additions: 1}

		require.NoError(t, run.Finish(PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunPlanned, run.status)
		require.Equal(t, PhaseFinished, run.plan.status)
		require.Equal(t, PhasePending, run.apply.status)
	})

	t.Run("finish plan with changes on run with autoapply enabled", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{
			Status:    RunPlanning,
			AutoApply: true,
		})

		run.plan.ResourceReport = &ResourceReport{Additions: 1}

		require.NoError(t, run.Finish(PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunApplyQueued, run.status)
		require.Equal(t, PhaseFinished, run.plan.status)
		require.Equal(t, PhaseQueued, run.apply.status)
	})

	t.Run("enqueue apply", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{Status: RunPlanned})

		require.NoError(t, run.EnqueueApply())

		require.Equal(t, RunApplyQueued, run.status)
		require.Equal(t, PhaseQueued, run.apply.status)
	})

	t.Run("start apply", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{Status: RunApplyQueued})

		require.NoError(t, run.Start(ApplyPhase))

		require.Equal(t, RunApplying, run.status)
		require.Equal(t, PhaseRunning, run.apply.status)
	})

	t.Run("finish apply", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{Status: RunApplying})

		require.NoError(t, run.Finish(ApplyPhase, PhaseFinishOptions{}))

		require.Equal(t, RunApplied, run.status)
		require.Equal(t, PhaseFinished, run.apply.status)
	})

	t.Run("finish apply with errors", func(t *testing.T) {
		run := NewTestRun(t, TestRunCreateOptions{Status: RunApplying})

		require.NoError(t, run.Finish(ApplyPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, RunErrored, run.status)
		require.Equal(t, PhaseErrored, run.apply.status)
	})
}

func TestRun_Cancel_Pending(t *testing.T) {
	run := NewTestRun(t, TestRunCreateOptions{Status: RunPending})
	enqueue, err := run.Cancel()
	require.NoError(t, err)
	assert.False(t, enqueue)
	assert.NotZero(t, run.forceCancelAvailableAt)
}

func TestRun_Cancel_Planning(t *testing.T) {
	run := NewTestRun(t, TestRunCreateOptions{Status: RunPlanning})
	enqueue, err := run.Cancel()
	require.NoError(t, err)
	assert.True(t, enqueue)
	assert.NotZero(t, run.forceCancelAvailableAt)
}
