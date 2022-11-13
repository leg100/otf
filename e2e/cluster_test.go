package e2e

import (
	"os/exec"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCluster is an end-to-end test of the clustering capabilities, i.e.
// more than one otfd, both connected to the same postgres db. The test runs two
// otfd daemons:
//
// otfd1) otfd to which the TF CLI connects
// otfd2) otfd to which the otf-agent connects
//
// This setup provides a limited demonstration that the cluster is co-ordinating
// processes successfully, e.g. relaying of logs from the agent through to the
// TF CLI
func TestCluster(t *testing.T) {
	addBuildsToPath(t)

	org := otf.NewTestOrganization(t)
	team := otf.NewTeam("owners", org)
	user := otf.NewTestUser(t, otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))

	// start two daemons
	daemon := &daemon{}
	daemon.withGithubUser(user)
	userDaemon := daemon.start(t)
	agentDaemon := daemon.start(t)

	// creating api token via web also syncs org
	_ = terraformLoginTasks(t, userDaemon)

	// org now sync'd, so we can create agent token via CLI
	agentToken := createAgentToken(t, org.Name(), userDaemon)
	// start agent, instructing it to connect to otfd2
	startAgent(t, agentToken, agentDaemon)

	// create root module, setting otfd1 as hostname
	root := newRootModule(t, userDaemon, org.Name(), "dev")

	// terraform init automatically creates a workspace named dev
	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// edit workspace to use agent
	cmd = exec.Command("otf", "workspaces", "edit", "dev", "--organization", org.Name(), "--execution-mode", "agent", "--address", userDaemon)
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	assert.Equal(t, "updated execution mode: agent\n", string(out))

	// terraform plan
	cmd = exec.Command("terraform", "plan", "-no-color")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")

	// terraform apply
	cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
}
