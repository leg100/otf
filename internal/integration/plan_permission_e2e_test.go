package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlanPermissionE2E demonstrates a user with plan permissions on a workspace interacting
// with the workspace via the terraform CLI.
func TestPlanPermissionE2E(t *testing.T) {
	t.Parallel()

	// Create user and org, and user becomes owner of the org
	svc := setup(t, nil)
	owner, ownerCtx := svc.createUserCtx(t, ctx)
	org := svc.createOrganization(t, ownerCtx)

	// Create engineer user and team and make member of a team
	engineer, engineerCtx := svc.createUserCtx(t, ctx)
	team := svc.createTeam(t, ownerCtx, org)
	err := svc.AddTeamMembership(ownerCtx, auth.TeamMembershipOptions{
		TeamID:   team.ID,
		Username: engineer.Username,
	})
	require.NoError(t, err)

	// create terraform configPath
	configPath := newRootModule(t, svc.Hostname(), org.Name, "my-test-workspace")

	// Open browser, using owner's credentials create workspace and assign plan
	// permissions to the engineer's team.
	browser := createBrowserCtx(t)
	err = chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ownerCtx, svc.Hostname(), owner.Username, svc.Secret),
		createWorkspace(t, svc.Hostname(), org.Name, "my-test-workspace"),
		addWorkspacePermission(t, svc.Hostname(), org.Name, "my-test-workspace", team.Name, "plan"),
	})
	require.NoError(t, err)

	// As engineer, run terraform init, and plan.
	_ = svc.tfcli(t, engineerCtx, "init", configPath)
	out := svc.tfcli(t, engineerCtx, "plan", configPath)
	assert.Contains(t, out, "Plan: 1 to add, 0 to change, 0 to destroy.")

	// Limited privileges should prohibit an apply
	out, err = svc.tfcliWithError(t, engineerCtx, "apply", configPath, "-auto-approve")
	if assert.Error(t, err) {
		assert.Contains(t, string(out), "Error: Insufficient rights to apply changes")
	}

	// Limited privileges should prohibit a destroy
	out, err = svc.tfcliWithError(t, engineerCtx, "destroy", configPath, "-auto-approve")
	if assert.Error(t, err) {
		assert.Contains(t, string(out), "Error: Insufficient rights to apply changes")
	}
}
