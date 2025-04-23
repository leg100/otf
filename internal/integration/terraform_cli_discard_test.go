package integration

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"
	"time"

	goexpect "github.com/google/goexpect"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TerraformCLIDiscard demonstrates a user discarding a run via
// the terraform CLI.
func TestIntegration_TerraformCLIDiscard(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

	// create some config and run terraform init
	configPath := newRootModule(t, svc.System.Hostname(), org.Name, t.Name())
	svc.engineCLI(t, ctx, "", "init", configPath)

	// Create user token expressly for terraform apply
	_, token := svc.createToken(t, ctx, nil)

	// Invoke terraform apply
	e, tferr, err := goexpect.SpawnWithArgs(
		[]string{terraform, "-chdir=" + configPath, "apply", "-no-color"},
		time.Minute,
		goexpect.PartialMatch(true),
		goexpect.SetEnv(internal.SafeAppend(sharedEnvs, internal.CredentialEnv(svc.System.Hostname(), token))),
	)
	require.NoError(t, err)
	defer e.Close()

	// Discard run
	e.ExpectBatch([]goexpect.Batcher{
		&goexpect.BExp{R: fmt.Sprintf(`Do you want to perform these actions in workspace "%s"`, t.Name())},
		&goexpect.BExp{R: "Enter a value:"}, &goexpect.BSnd{S: "no\n"},
		&goexpect.BExp{R: "Error: Apply discarded."},
	}, time.Minute)

	var exitError *exec.ExitError
	require.True(t, errors.As(<-tferr, &exitError))
	require.Equal(t, 1, exitError.ExitCode())

	runs, err := svc.Runs.List(ctx, run.ListOptions{Organization: &org.Name})
	require.NoError(t, err)
	require.Equal(t, 1, len(runs.Items))
	require.Equal(t, runstatus.Discarded, runs.Items[0].Status)
}
