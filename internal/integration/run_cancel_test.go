package integration

import (
	"testing"
	"time"

	expect "github.com/google/goexpect"
	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunCancel demonstrates a run being canceled mid-flow. The
// agent should terminate
func TestIntegration_RunCancel(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	ws := svc.createWorkspace(t, ctx, nil)
	cv := svc.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	run := svc.createRun(t, ctx, ws, cv)

	svc.GetLogs
	// Invoke terraform plan
	_, token := svc.createToken(t, ctx, nil)
	e, tferr, err := expect.SpawnWithArgs(
		[]string{"terraform", "-chdir=" + config, "plan", "-no-color"},
		time.Minute,
		expect.PartialMatch(true),
		expect.SetEnv(
			append(envs, internal.CredentialEnv(svc.Hostname(), token)),
		),
	)
	require.NoError(t, err)
	defer e.Close()

	// wait for terraform plan to call handler
	<-planning
	svc.Cancel(ctx, run.ID)
	close(interrupted)

	// Confirm canceling run
	e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: "Do you want to cancel the remote operation?"},
		&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: "yes\n"},
		&expect.BExp{R: "The remote operation was successfully cancelled."},
	}, time.Minute)
	// Terraform should return with exit code 0
	require.NoError(t, <-tferr)

	runs, err := svc.ListRuns(ctx, run.RunListOptions{Organization: &org.Name})
	require.NoError(t, err)
	require.Equal(t, 1, len(runs.Items))
	require.Equal(t, internal.RunCanceled, runs.Items[0].Status)
}
