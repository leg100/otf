package e2e

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSandbox demonstrates the sandbox feature, whereby terraform apply is run
// within an isolated environment.
func TestSandbox(t *testing.T) {
	if _, err := exec.LookPath("bwrap"); errors.Is(err, exec.ErrNotFound) {
		t.Skipf("bwrap binary not found")
	}

	addBuildsToPath(t)

	user := otf.NewTestUser(t)
	hostname := startDaemon(t, user, "--sandbox")

	// create root module using user's personal organization
	root := newRootModule(t, hostname, user.Username(), "dev")

	userToken := createAPIToken(t, hostname)
	login(t, hostname, userToken)

	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// terraform apply
	cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	assert.Contains(t, string(out), "Running within sandbox...")
	assert.Contains(t, string(out), "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
}
