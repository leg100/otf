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

	// Build and start a daemon specifically for the boss
	bossDaemon := &daemon{}
	bossDaemon.withGithubUser(boss)
	hostname := bossDaemon.start(t)
	url := "https://" + hostname

	// create workspace via web - note this also syncs the org and owner
	allocater := newBrowserAllocater(t)
	workspaceName, workspaceID := createWebWorkspace(t, allocater, url, org.Name())

	// assign plan permissions to devops team
	addWorkspacePermission(t, allocater, url, org.Name(), workspaceID, devops.Name(), "plan")

	// setup non-owner engineer user - note we start another daemon because this is the
	// only way at present that an additional user can be seeded for testing.
	engineer := otf.NewUser("engineer", otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(devops))
	engineerDaemon := &daemon{}
	engineerDaemon.withGithubUser(engineer)
	engineerHostname := engineerDaemon.start(t)

	_ = terraformLoginTasks(t, engineerHostname)

	root := newRootModule(t, engineerHostname, org.Name(), workspaceName)

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
