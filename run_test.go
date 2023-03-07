package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_States(t *testing.T) {
	t.Run("pending", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})

		require.Equal(t, RunPending, run.Status)
		require.Equal(t, PhasePending, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})

		require.NoError(t, run.EnqueuePlan())

		require.Equal(t, RunPlanQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("start plan", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
		run.Status = RunPlanQueued

		require.NoError(t, run.Start(PlanPhase))

		require.Equal(t, RunPlanning, run.Status)
		require.Equal(t, PhaseRunning, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
		run.Status = RunPlanning

		require.NoError(t, run.Finish(PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunPlannedAndFinished, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with errors", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
		run.Status = RunPlanning

		require.NoError(t, run.Finish(PlanPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, RunErrored, run.Status)
		require.Equal(t, PhaseErrored, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with changes", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
		run.Status = RunPlanning

		run.Plan.ResourceReport = &ResourceReport{Additions: 1}

		require.NoError(t, run.Finish(PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunPlanned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with changes on run with autoapply enabled", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{
			AutoApply: Bool(true),
		})
		run.Status = RunPlanning

		run.Plan.ResourceReport = &ResourceReport{Additions: 1}

		require.NoError(t, run.Finish(PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunApplyQueued, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("enqueue apply", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
		run.Status = RunPlanned

		require.NoError(t, run.EnqueueApply())

		require.Equal(t, RunApplyQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("start apply", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
		run.Status = RunApplyQueued

		require.NoError(t, run.Start(ApplyPhase))

		require.Equal(t, RunApplying, run.Status)
		require.Equal(t, PhaseRunning, run.Apply.Status)
	})

	t.Run("finish apply", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
		run.Status = RunApplying

		require.NoError(t, run.Finish(ApplyPhase, PhaseFinishOptions{}))

		require.Equal(t, RunApplied, run.Status)
		require.Equal(t, PhaseFinished, run.Apply.Status)
	})

	t.Run("finish apply with errors", func(t *testing.T) {
		run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
		run.Status = RunApplying

		require.NoError(t, run.Finish(ApplyPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, RunErrored, run.Status)
		require.Equal(t, PhaseErrored, run.Apply.Status)
	})
}

func TestRun_Cancel_Pending(t *testing.T) {
	run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
	enqueue, err := run.Cancel()
	require.NoError(t, err)
	assert.False(t, enqueue)
	assert.NotZero(t, run.ForceCancelAvailableAt)
}

func TestRun_Cancel_Planning(t *testing.T) {
	run := NewRun(&ConfigurationVersion{}, &Workspace{}, RunCreateOptions{})
	run.Status = RunPlanning
	enqueue, err := run.Cancel()
	require.NoError(t, err)
	assert.True(t, enqueue)
	assert.NotZero(t, run.ForceCancelAvailableAt)
}
