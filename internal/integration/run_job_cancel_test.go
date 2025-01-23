package integration

import (
	"testing"

	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunJobCancel tests the cancelation of a run and one of its
// jobs, when the job is not yet running.
func TestIntegration_RunJobCancel(t *testing.T) {
	integrationTest(t)

	// Disable runner to prevent the run's plan job from running.
	daemon, _, ctx := setup(t, &config{
		Config: daemon.Config{
			DisableRunner: true,
		},
	})
	// Watch job events
	jobs, unsub := daemon.Runners.WatchJobs(ctx)
	defer unsub()

	// Create run, and wait til it reaches plan queued state
	r := daemon.createRun(t, ctx, nil, nil, nil)
	daemon.waitRunStatus(t, r.ID, run.RunPlanQueued)
	// Job should be automatically created
	wait(t, jobs, func(event pubsub.Event[*runner.Job]) bool {
		return event.Payload.RunID == r.ID
	})

	// Cancel run
	err := daemon.Runs.Cancel(ctx, r.ID)
	require.NoError(t, err)

	// Run and job should now enter canceled state.
	daemon.waitRunStatus(t, r.ID, run.RunCanceled)
	wait(t, jobs, func(event pubsub.Event[*runner.Job]) bool {
		return event.Payload.Status == runner.JobCanceled &&
			event.Payload.RunID == r.ID
	})
}
