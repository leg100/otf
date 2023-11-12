package run

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_New_CreatedBy(t *testing.T) {
	ctx := context.Background()
	ctx = internal.AddSubjectToContext(ctx, &user.User{Username: "terry"})
	run := newTestRun(ctx, CreateOptions{})
	assert.NotNil(t, run.CreatedBy)
	assert.Equal(t, "terry", *run.CreatedBy)
}

func TestRun_States(t *testing.T) {
	ctx := context.Background()

	t.Run("pending", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})

		require.Equal(t, RunPending, run.Status)
		require.Equal(t, PhasePending, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})

		require.NoError(t, run.EnqueuePlan())

		require.Equal(t, RunPlanQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("start plan", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = RunPlanQueued

		require.NoError(t, run.Start(internal.PlanPhase))

		require.Equal(t, RunPlanning, run.Status)
		require.Equal(t, PhaseRunning, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = RunPlanning

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunPlannedAndFinished, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with errors", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = RunPlanning

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, RunErrored, run.Status)
		require.Equal(t, PhaseErrored, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with resource changes", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = RunPlanning

		run.Plan.ResourceReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunPlanned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with output changes", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = RunPlanning

		run.Plan.OutputReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunPlanned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with changes on run with autoapply enabled", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{
			AutoApply: internal.Bool(true),
		})
		run.Status = RunPlanning

		run.Plan.ResourceReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunApplyQueued, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("finish plan with cost estimation enabled", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.CostEstimationEnabled = true
		run.Status = RunPlanning

		run.Plan.ResourceReport = &Report{Additions: 1}

		require.NoError(t, run.Finish(internal.PlanPhase, PhaseFinishOptions{}))

		require.Equal(t, RunCostEstimated, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("enqueue apply", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = RunPlanned

		require.NoError(t, run.EnqueueApply())

		require.Equal(t, RunApplyQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("start apply", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = RunApplyQueued

		require.NoError(t, run.Start(internal.ApplyPhase))

		require.Equal(t, RunApplying, run.Status)
		require.Equal(t, PhaseRunning, run.Apply.Status)
	})

	t.Run("finish apply", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = RunApplying

		require.NoError(t, run.Finish(internal.ApplyPhase, PhaseFinishOptions{}))

		require.Equal(t, RunApplied, run.Status)
		require.Equal(t, PhaseFinished, run.Apply.Status)
	})

	t.Run("finish apply with errors", func(t *testing.T) {
		run := newTestRun(ctx, CreateOptions{})
		run.Status = RunApplying

		require.NoError(t, run.Finish(internal.ApplyPhase, PhaseFinishOptions{Errored: true}))

		require.Equal(t, RunErrored, run.Status)
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
	var (
		now = internal.CurrentTimestamp(nil)
		ago = func(seconds int) time.Time {
			return now.Add(time.Duration(seconds) * -time.Second)
		}
		createRun = func(created time.Time) *Run {
			return newRun(context.Background(), &organization.Organization{}, &configversion.ConfigurationVersion{}, &workspace.Workspace{}, CreateOptions{now: &created})
		}
	)

	tests := []struct {
		name string
		run  func() *Run
		want []StatusPeriod
	}{
		{
			"fresh run",
			func() *Run { return createRun(ago(0)) },
			[]StatusPeriod{
				{Status: RunPending, Period: 0},
			},
		},
		{
			"planning",
			func() *Run {
				// 1 second in pending state
				// 1 second in plan queued state
				// 2 seconds in planning state
				return createRun(ago(4)).
					updateStatus(RunPlanQueued, internal.Time(ago(3))).
					updateStatus(RunPlanning, internal.Time(ago(2)))
			},
			[]StatusPeriod{
				{Status: RunPending, Period: time.Second},
				{Status: RunPlanQueued, Period: time.Second},
				{Status: RunPlanning, Period: 2 * time.Second},
			},
		},
		{
			"planned and finished",
			func() *Run {
				// 1 second in pending state
				// 1 second in plan queued state
				// 2 seconds in planning state
				// finished
				return createRun(ago(4)).
					updateStatus(RunPlanQueued, internal.Time(ago(3))).
					updateStatus(RunPlanning, internal.Time(ago(2))).
					updateStatus(RunPlannedAndFinished, &now)
			},
			[]StatusPeriod{
				{Status: RunPending, Period: time.Second},
				{Status: RunPlanQueued, Period: time.Second},
				{Status: RunPlanning, Period: 2 * time.Second},
			},
		},
		{
			"applied",
			func() *Run {
				// 1 second in pending state
				// 1 second in plan queued state
				// 2 seconds in planning state
				// 1 seconds in planned state
				// 5 second in applying state
				// finished
				return createRun(ago(10)).
					updateStatus(RunPlanQueued, internal.Time(ago(9))).
					updateStatus(RunPlanning, internal.Time(ago(8))).
					updateStatus(RunPlanned, internal.Time(ago(6))).
					updateStatus(RunApplying, internal.Time(ago(5))).
					updateStatus(RunPlannedAndFinished, &now)
			},
			[]StatusPeriod{
				{Status: RunPending, Period: time.Second},
				{Status: RunPlanQueued, Period: time.Second},
				{Status: RunPlanning, Period: 2 * time.Second},
				{Status: RunPlanned, Period: time.Second},
				{Status: RunApplying, Period: 5 * time.Second},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.run().PeriodReport(now)
			assert.Equal(t, tt.want, got.Periods)
		})
	}
}

func newTestRun(ctx context.Context, opts CreateOptions) *Run {
	return newRun(ctx, &organization.Organization{}, &configversion.ConfigurationVersion{}, &workspace.Workspace{}, opts)
}
