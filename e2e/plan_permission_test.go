package e2e

import (
	"os/exec"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlanPermission demonstrates a user with plan permissions on a workspace interacting
// with the workspace via the terraform CLI.
func TestPlanPermission(t *testing.T) {
	addBuildsToPath(t)

	// First we need to setup an organization with a user who is both in the
	// owners team and the devops team.
	org := otf.NewTestOrganization(t)
	owners := otf.NewTeam("owners", org)
	devops := otf.NewTeam("devops", org)
	boss := otf.NewUser("boss", otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(owners, devops))
	hostname := startDaemon(t, boss)
	url := "https://" + hostname

	// create workspace via web - note this also syncs the org and owner
	allocater := newBrowserAllocater(t)
	workspace := createWebWorkspace(t, allocater, url, org.Name())

	// assign plan permissions to devops team
	addWorkspacePermission(t, allocater, url, org.Name(), workspace, devops.Name(), "plan")

	// setup non-owner user - note we start another daemon because this is the
	// only way at present that an additional user can be seeded for testing.
	engineer := otf.NewUser("engineer", otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(devops))
	hostname = startDaemon(t, engineer)

	engineerToken := createAPIToken(t, hostname)
	login(t, hostname, engineerToken)

	root := newRootModule(t, hostname, org.Name(), workspace)

	// terraform init
	cmd := exec.Command("terraform", "init", "-no-color")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	// terraform plan
	cmd = exec.Command("terraform", "plan", "-no-color")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	assert.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")

	// terraform apply
	cmd = exec.Command("terraform", "apply", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	if assert.Error(t, err) {
		assert.Contains(t, string(out), "Error: Insufficient rights to apply changes")
	}

	// terraform destroy
	cmd = exec.Command("terraform", "destroy", "-no-color", "-auto-approve")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	if assert.Error(t, err) {
		assert.Contains(t, string(out), "Error: Insufficient rights to apply changes")
	}
}
