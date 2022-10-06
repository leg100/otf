package e2e

import (
	"net/url"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCluster is an end-to-end test of the clustering capabilities, i.e.
// running more than one otfd. The test runs two otfd's (the first of which is
// expected to already been running):
//
// (1) otfd to which the TF CLI connects
// (2) otfd to which the otf-agent connects
//
// This setup demonstrates that the cluster can coordinate processes between
// the two clients.
func TestCluster(t *testing.T) {
	tfpath := terraformPath(t)
	t.Run("login", login(t, tfpath))
	organization := createOrganization(t)
	token := createAgentToken(t, organization)
	root := newRoot(t, organization)

	daemonURL := startDaemon(t)
	u, err := url.Parse(daemonURL)
	require.NoError(t, err)

	startAgent(t, token, u.Host)

	// terraform init creates a workspace named dev
	t.Run("terraform init", func(t *testing.T) {
		chdir(t, root)
		cmd := exec.Command(tfpath, "init", "-no-color")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
	})

	t.Run("edit workspace to use agent", func(t *testing.T) {
		cmd := exec.Command(client, "workspaces", "edit", "dev", "--organization", organization, "--execution-mode", "agent")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
		assert.Equal(t, "updated execution mode: agent\n", string(out))
	})

	t.Run("terraform plan", func(t *testing.T) {
		chdir(t, root)
		cmd := exec.Command(tfpath, "plan", "-no-color")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
		require.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")
	})

	t.Run("terraform apply", func(t *testing.T) {
		chdir(t, root)
		cmd := exec.Command(tfpath, "apply", "-no-color", "-auto-approve")
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
		require.Contains(t, string(out), "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
	})
}
