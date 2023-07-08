package run

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_States(t *testing.T) {
	t.Run("pending", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})

		require.Equal(t, internal.RunPending, run.Status)
		require.Equal(t, PhasePending, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})

		require.NoError(t, run.EnqueuePlan())

		require.Equal(t, internal.RunPlanQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("start plan", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = internal.RunPlanQueued

		require.NoError(t, run.Start(internal.PlanPhase))

		require.Equal(t, internal.RunPlanning, run.Status)
		require.Equal(t, PhaseRunning, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = internal.RunPlanning

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunPlannedAndFinished, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with errors", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = internal.RunPlanning

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, internal.RunErrored, run.Status)
		require.Equal(t, PhaseErrored, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with resource changes", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = internal.RunPlanning

		run.Plan.ResourceReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunPlanned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with output changes", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = internal.RunPlanning

		run.Plan.OutputReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunPlanned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with changes on run with autoapply enabled", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{
			AutoApply: internal.Bool(true),
		})
		run.Status = internal.RunPlanning

		run.Plan.ResourceReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunApplyQueued, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("enqueue apply", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = internal.RunPlanned

		require.NoError(t, run.EnqueueApply())

		require.Equal(t, internal.RunApplyQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("start apply", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = internal.RunApplyQueued

		require.NoError(t, run.Start(internal.ApplyPhase))

		require.Equal(t, internal.RunApplying, run.Status)
		require.Equal(t, PhaseRunning, run.Apply.Status)
	})

	t.Run("finish apply", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = internal.RunApplying

		require.NoError(t, run.Finish(internal.ApplyPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunApplied, run.Status)
		require.Equal(t, PhaseFinished, run.Apply.Status)
	})

	t.Run("finish apply with errors", func(t *testing.T) {
		run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
		run.Status = internal.RunApplying

		require.NoError(t, run.Finish(internal.ApplyPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, internal.RunErrored, run.Status)
		require.Equal(t, PhaseErrored, run.Apply.Status)
	})
}

func TestRun_Cancel(t *testing.T) {
	run := newRun(&configversion.ConfigurationVersion{}, &workspace.Workspace{}, RunCreateOptions{})
	err := run.Cancel()
	require.NoError(t, err)
	assert.NotZero(t, run.ForceCancelAvailableAt)
}
