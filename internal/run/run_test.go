package run

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_New_CreatedBy(t *testing.T) {
	ctx := context.Background()
	ctx = internal.AddSubjectToContext(ctx, &auth.User{Username: "terry"})
	run := newTestRun(ctx, CreateOptions{})
	assert.NotNil(t, run.CreatedBy)
	assert.Equal(t, "terry", *run.CreatedBy)
}

func TestRun_States(t *testing.T) {
	ctx := context.Background()

	t.Run("pending", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})

		require.Equal(t, internal.RunPending, run.Status)
		require.Equal(t, PhasePending, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})

		require.NoError(t, run.EnqueuePlan())

		require.Equal(t, internal.RunPlanQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("start plan", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = internal.RunPlanQueued

		require.NoError(t, run.Start(internal.PlanPhase))

		require.Equal(t, internal.RunPlanning, run.Status)
		require.Equal(t, PhaseRunning, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = internal.RunPlanning

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunPlannedAndFinished, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with errors", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = internal.RunPlanning

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, internal.RunErrored, run.Status)
		require.Equal(t, PhaseErrored, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with resource changes", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = internal.RunPlanning

		run.Plan.ResourceReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunPlanned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with output changes", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = internal.RunPlanning

		run.Plan.OutputReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunPlanned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with changes on run with autoapply enabled", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{
			AutoApply: internal.Bool(true),
		})
		run.Status = internal.RunPlanning

		run.Plan.ResourceReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunApplyQueued, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("finish plan with cost estimation enabled", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.CostEstimationEnabled = true
		run.Status = internal.RunPlanning

		run.Plan.ResourceReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunCostEstimated, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("enqueue apply", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = internal.RunPlanned

		require.NoError(t, run.EnqueueApply())

		require.Equal(t, internal.RunApplyQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("start apply", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = internal.RunApplyQueued

		require.NoError(t, run.Start(internal.ApplyPhase))

		require.Equal(t, internal.RunApplying, run.Status)
		require.Equal(t, PhaseRunning, run.Apply.Status)
	})

	t.Run("finish apply", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = internal.RunApplying

		require.NoError(t, run.Finish(internal.ApplyPhase, PhaseFinishOptions{}))

		require.Equal(t, internal.RunApplied, run.Status)
		require.Equal(t, PhaseFinished, run.Apply.Status)
	})

	t.Run("finish apply with errors", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = internal.RunApplying

		require.NoError(t, run.Finish(internal.ApplyPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, internal.RunErrored, run.Status)
		require.Equal(t, PhaseErrored, run.Apply.Status)
	})

	t.Run("cancel run", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		err := run.Cancel()
		require.NoError(t, err)
		assert.NotZero(t, run.ForceCancelAvailableAt)
	})
}

func TestRun_StatusReport(t *testing.T) {
	now := internal.CurrentTimestamp()
	ago := func(seconds int) time.Time {
		return now.Add(time.Duration(seconds) * -time.Second)
	}

	// TODO: handle 0 time
	tests := []struct {
		name       string
		timestamps []StatusTimestamp
		want       map[internal.RunStatus]int
	}{
		{
			"fresh run",
			[]StatusTimestamp{{Status: internal.RunPending, Timestamp: now.Add(-time.Second)}},
			map[internal.RunStatus]int{
				internal.RunPending: 100,
			},
		},
		{
			"planning",
			[]StatusTimestamp{
				// 1 second in pending state
				{Status: internal.RunPending, Timestamp: now.Add(4 * -time.Second)},
				// 1 second in plan queued state
				{Status: internal.RunPlanQueued, Timestamp: now.Add(3 * -time.Second)},
				// 2 seconds in planning state
				{Status: internal.RunPlanning, Timestamp: now.Add(2 * -time.Second)},
			},
			map[internal.RunStatus]int{
				internal.RunPending:    25,
				internal.RunPlanQueued: 25,
				internal.RunPlanning:   50,
			},
		},
		{
			"planned and finished",
			[]StatusTimestamp{
				// 1 second in pending state
				{Status: internal.RunPending, Timestamp: now.Add(4 * -time.Second)},
				// 1 second in plan queued state
				{Status: internal.RunPlanQueued, Timestamp: now.Add(3 * -time.Second)},
				// 2 seconds in planning state
				{Status: internal.RunPlanning, Timestamp: now.Add(2 * -time.Second)},
				// finished
				{Status: internal.RunPlannedAndFinished, Timestamp: now},
			},
			map[internal.RunStatus]int{
				internal.RunPending:    25,
				internal.RunPlanQueued: 25,
				internal.RunPlanning:   50,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := &Run{StatusTimestamps: tt.timestamps}
			got := run.StatusReport(now)
			assert.Equal(t, tt.want, got)
		})
	}
}

func newTestRun(ctx context.Context, opts CreateOptions) *Run {
	return newRun(ctx, &organization.Organization{}, &configversion.ConfigurationVersion{}, &workspace.Workspace{}, opts)
}
