package integration

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSandbox demonstrates the sandbox feature, whereby terraform apply is run
// within an isolated environment.
func TestSandbox(t *testing.T) {
	integrationTest(t)

	_, err := exec.LookPath("bwrap")
	if errors.Is(err, exec.ErrNotFound) {
		t.Skip("install bwrap before running this test")
	}
	require.NoError(t, err)

	daemon, org, ctx := setup(t, withServerRunnerDebug(), withServerRunnerSandbox())

	// create terraform config
	config := newRootModule(t, daemon.System.Hostname(), org.Name, "dev")
	// terraform init
	daemon.engineCLI(t, ctx, "", "init", config)
	// terraform apply
	out := daemon.engineCLI(t, ctx, "", "apply", config, "-auto-approve")
	assert.Contains(t, out, "Sandbox mode: true")
	assert.Contains(t, out, "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
}
