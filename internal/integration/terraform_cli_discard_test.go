package integration

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"

	expect "github.com/google/goexpect"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TerraformCLIDiscard demonstrates a user discarding a run via
// the terraform CLI.
func TestIntegration_TerraformCLIDiscard(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// create some config and run terraform init
	configPath := newRootModule(t, svc.Hostname(), org.Name, t.Name())
	svc.tfcli(t, ctx, "init", configPath)

	// Create user token expressly for terraform apply
	_, token := svc.createToken(t, ctx, nil)

	// Invoke terraform apply
	e, tferr, err := expect.SpawnWithArgs(
		[]string{"terraform", "-chdir=" + configPath, "apply", "-no-color"},
		time.Minute,
		expect.PartialMatch(true),
		expect.SetEnv(internal.SafeAppend(sharedEnvs, internal.CredentialEnv(svc.Hostname(), token))),
	)
	require.NoError(t, err)
	defer e.Close()

	// Discard run
	e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: fmt.Sprintf(`Do you want to perform these actions in workspace "%s"`, t.Name())},
		&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: "no\n"},
		&expect.BExp{R: "Error: Apply discarded."},
	}, time.Minute)

	var exitError *exec.ExitError
	require.True(t, errors.As(<-tferr, &exitError))
	require.Equal(t, 1, exitError.ExitCode())

	runs, err := svc.ListRuns(ctx, run.RunListOptions{Organization: &org.Name})
	require.NoError(t, err)
	require.Equal(t, 1, len(runs.Items))
	require.Equal(t, internal.RunDiscarded, runs.Items[0].Status)
}
