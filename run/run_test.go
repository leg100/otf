package run

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_States(t *testing.T) {
	t.Run("pending", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})

		require.Equal(t, otf.RunPending, run.Status)
		require.Equal(t, PhasePending, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})

		require.NoError(t, run.EnqueuePlan())

		require.Equal(t, otf.RunPlanQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("start plan", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = otf.RunPlanQueued

		require.NoError(t, run.Start(otf.PlanPhase))

		require.Equal(t, otf.RunPlanning, run.Status)
		require.Equal(t, PhaseRunning, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = otf.RunPlanning

		require.NoError(t, run.Finish(otf.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, otf.RunPlannedAndFinished, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with errors", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = otf.RunPlanning

		require.NoError(t, run.Finish(otf.PlanPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, otf.RunErrored, run.Status)
		require.Equal(t, PhaseErrored, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with changes", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = otf.RunPlanning

		run.Plan.ResourceReport = &ResourceReport{Additions: 1}

		require.NoError(t, run.Finish(otf.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, otf.RunPlanned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with changes on run with autoapply enabled", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{
			AutoApply: otf.Bool(true),
		})
		run.Status = otf.RunPlanning

		run.Plan.ResourceReport = &ResourceReport{Additions: 1}

		require.NoError(t, run.Finish(otf.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, otf.RunApplyQueued, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("enqueue apply", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = otf.RunPlanned

		require.NoError(t, run.EnqueueApply())

		require.Equal(t, otf.RunApplyQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("start apply", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = otf.RunApplyQueued

		require.NoError(t, run.Start(otf.ApplyPhase))

		require.Equal(t, otf.RunApplying, run.Status)
		require.Equal(t, PhaseRunning, run.Apply.Status)
	})

	t.Run("finish apply", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = otf.RunApplying

		require.NoError(t, run.Finish(otf.ApplyPhase, PhaseFinishOptions{}))

		require.Equal(t, otf.RunApplied, run.Status)
		require.Equal(t, PhaseFinished, run.Apply.Status)
	})

	t.Run("finish apply with errors", func(t *testing.T) {
		run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = otf.RunApplying

		require.NoError(t, run.Finish(otf.ApplyPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, otf.RunErrored, run.Status)
		require.Equal(t, PhaseErrored, run.Apply.Status)
	})
}

func TestRun_Cancel_Pending(t *testing.T) {
	run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
	enqueue, err := run.Cancel()
	require.NoError(t, err)
	assert.False(t, enqueue)
	assert.NotZero(t, run.ForceCancelAvailableAt)
}

func TestRun_Cancel_Planning(t *testing.T) {
	run := NewRun(&otf.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
	run.Status = otf.RunPlanning
	enqueue, err := run.Cancel()
	require.NoError(t, err)
	assert.True(t, enqueue)
	assert.NotZero(t, run.ForceCancelAvailableAt)
}
