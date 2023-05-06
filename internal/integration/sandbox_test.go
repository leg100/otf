package integration

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSandbox demonstrates the sandbox feature, whereby terraform apply is run
// within an isolated environment.
func TestSandbox(t *testing.T) {
	_, err := exec.LookPath("bwrap")
	if errors.Is(err, exec.ErrNotFound) {
		t.Skip("install bwrap before running this test")
	}
	require.NoError(t, err)

	daemon := setup(t, &config{Config: daemon.Config{
		AgentConfig: &agent.Config{
			Sandbox: true,
			Debug:   true,
		},
	}})
	_, ctx := daemon.createUserCtx(t, ctx)
	org := daemon.createOrganization(t, ctx)

	// create terraform config
	config := newRootModule(t, daemon.Hostname(), org.Name, "dev")
	// terraform init
	daemon.tfcli(t, ctx, "init", config)
	// terraform apply
	out := daemon.tfcli(t, ctx, "apply", config, "-auto-approve")
	assert.Contains(t, out, "Sandbox mode: true")
	assert.Contains(t, out, "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
}
