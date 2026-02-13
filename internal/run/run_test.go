package run

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_New_CreatedBy(t *testing.T) {
	ctx := context.Background()
	user := user.NewTestUser(t)
	ctx = authz.AddSubjectToContext(ctx, user)
	run := newTestRun(t, ctx, CreateOptions{})
	assert.NotNil(t, run.CreatedBy)
	assert.Equal(t, user.Username, *run.CreatedBy)
}

func TestRun_States(t *testing.T) {
	ctx := context.Background()

	t.Run("pending", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})

		require.Equal(t, runstatus.Pending, run.Status)
		require.Equal(t, PhasePending, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})

		require.NoError(t, run.EnqueuePlan())

		require.Equal(t, runstatus.PlanQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("start plan", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.PlanQueued

		require.NoError(t, run.Start())

		require.Equal(t, runstatus.Planning, run.Status)
		require.Equal(t, PhaseRunning, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning

		_, err := run.Finish(PlanPhase, PhaseFinishOptions{})
		require.NoError(t, err)

		require.Equal(t, runstatus.PlannedAndFinished, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with errors", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning

		_, err := run.Finish(PlanPhase, PhaseFinishOptions{Errored: true})
		require.NoError(t, err)

		require.Equal(t, runstatus.Errored, run.Status)
		require.Equal(t, PhaseErrored, run.Plan.Status)
		require.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("finish plan with resource changes", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning

		run.Plan.ResourceReport = &Report{Additions: 1}

		_, err := run.Finish(PlanPhase, PhaseFinishOptions{})
		require.NoError(t, err)

		require.Equal(t, runstatus.Planned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with output changes", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning

		run.Plan.OutputReport = &Report{Additions: 1}

		_, err := run.Finish(PlanPhase, PhaseFinishOptions{})
		require.NoError(t, err)

		require.Equal(t, runstatus.Planned, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with changes on run with autoapply enabled", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{AutoApply: new(true)})
		run.Status = runstatus.Planning

		run.Plan.ResourceReport = &Report{Additions: 1}

		autoapply, err := run.Finish(PlanPhase, PhaseFinishOptions{})
		require.NoError(t, err)

		assert.True(t, autoapply)
		assert.Equal(t, runstatus.Planned, run.Status)
		assert.Equal(t, PhaseFinished, run.Plan.Status)
		assert.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("finish plan with cost estimation enabled", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.CostEstimationEnabled = true
		run.Status = runstatus.Planning

		run.Plan.ResourceReport = &Report{Additions: 1}

		_, err := run.Finish(PlanPhase, PhaseFinishOptions{})
		require.NoError(t, err)

		require.Equal(t, runstatus.CostEstimated, run.Status)
		require.Equal(t, PhaseFinished, run.Plan.Status)
		require.Equal(t, PhasePending, run.Apply.Status)
	})

	t.Run("enqueue apply", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planned

		require.NoError(t, run.EnqueueApply())

		require.Equal(t, runstatus.ApplyQueued, run.Status)
		require.Equal(t, PhaseQueued, run.Apply.Status)
	})

	t.Run("start apply", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.ApplyQueued

		require.NoError(t, run.Start())

		require.Equal(t, runstatus.Applying, run.Status)
		require.Equal(t, PhaseRunning, run.Apply.Status)
	})

	t.Run("finish apply", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Applying

		_, err := run.Finish(ApplyPhase, PhaseFinishOptions{})
		require.NoError(t, err)

		require.Equal(t, runstatus.Applied, run.Status)
		require.Equal(t, PhaseFinished, run.Apply.Status)
	})

	t.Run("finish apply with errors", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Applying

		_, err := run.Finish(ApplyPhase, PhaseFinishOptions{Errored: true})
		require.NoError(t, err)

		require.Equal(t, runstatus.Errored, run.Status)
		require.Equal(t, PhaseErrored, run.Apply.Status)
	})

	t.Run("cancel pending run", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		err := run.Cancel(true, false)
		require.NoError(t, err)
		// no signal should be sent
		assert.Zero(t, run.CancelSignaledAt)
		assert.Equal(t, PhaseUnreachable, run.Plan.Status)
		assert.Equal(t, PhaseUnreachable, run.Apply.Status)
	})

	t.Run("cancel planning run should indicate signal be sent", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning
		err := run.Cancel(true, false)
		require.NoError(t, err)
		assert.NotZero(t, run.CancelSignaledAt)
		assert.Equal(t, runstatus.Planning, run.Status)
	})

	t.Run("when non-user cancels a planning run, it should be placed into canceled state", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning
		err := run.Cancel(false, false)
		require.NoError(t, err)
		assert.Equal(t, PhaseCanceled, run.Plan.Status)
		assert.Equal(t, PhaseUnreachable, run.Apply.Status)
		assert.Equal(t, runstatus.Canceled, run.Status)
	})

	t.Run("user cannot cancel a run twice", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning
		err := run.Cancel(true, false)
		require.NoError(t, err)
		err = run.Cancel(true, false)
		assert.Equal(t, ErrRunCancelNotAllowed, err)
	})

	t.Run("cannot force cancel a run when no previous attempt has been made to cancel run gracefully", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning
		err := run.Cancel(true, true)
		assert.Equal(t, ErrRunForceCancelNotAllowed, err)
	})

	t.Run("force cancel run when graceful cancel has already been attempted and cool off period has elapsed", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning
		// gracefully canceled 11 seconds ago
		run.CancelSignaledAt = new(time.Now().Add(-11 * time.Second))
		// force cancel now
		err := run.Cancel(true, true)
		require.NoError(t, err)
		assert.Equal(t, runstatus.ForceCanceled, run.Status)
	})

	t.Run("non-user cannot force cancel a run", func(t *testing.T) {
		run := newTestRun(t, ctx, CreateOptions{})
		run.Status = runstatus.Planning
		err := run.Cancel(false, true)
		assert.Equal(t, ErrRunForceCancelNotAllowed, err)
	})
}

func TestRun_StatusReport(t *testing.T) {
	var (
		now = internal.CurrentTimestamp(nil)
		ago = func(seconds int) time.Time {
			return now.Add(time.Duration(seconds) * -time.Second)
		}
		createRun = func(created time.Time) *Run {
			return newTestRun(t, t.Context(), CreateOptions{
				now: &created,
			})
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
				{Status: runstatus.Pending, Period: 0},
			},
		},
		{
			"planning",
			func() *Run {
				// 1 second in pending state
				// 1 second in plan queued state
				// 2 seconds in planning state
				return createRun(ago(4)).
					updateStatus(runstatus.PlanQueued, new(ago(3))).
					updateStatus(runstatus.Planning, new(ago(2)))
			},
			[]StatusPeriod{
				{Status: runstatus.Pending, Period: time.Second},
				{Status: runstatus.PlanQueued, Period: time.Second},
				{Status: runstatus.Planning, Period: 2 * time.Second},
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
					updateStatus(runstatus.PlanQueued, new(ago(3))).
					updateStatus(runstatus.Planning, new(ago(2))).
					updateStatus(runstatus.PlannedAndFinished, &now)
			},
			[]StatusPeriod{
				{Status: runstatus.Pending, Period: time.Second},
				{Status: runstatus.PlanQueued, Period: time.Second},
				{Status: runstatus.Planning, Period: 2 * time.Second},
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
					updateStatus(runstatus.PlanQueued, new(ago(9))).
					updateStatus(runstatus.Planning, new(ago(8))).
					updateStatus(runstatus.Planned, new(ago(6))).
					updateStatus(runstatus.Applying, new(ago(5))).
					updateStatus(runstatus.PlannedAndFinished, &now)
			},
			[]StatusPeriod{
				{Status: runstatus.Pending, Period: time.Second},
				{Status: runstatus.PlanQueued, Period: time.Second},
				{Status: runstatus.Planning, Period: 2 * time.Second},
				{Status: runstatus.Planned, Period: time.Second},
				{Status: runstatus.Applying, Period: 5 * time.Second},
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
