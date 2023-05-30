package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_PlanPermission demonstrates the assignment of the workspace
// 'plan' role to a team and what they can and cannot do with that role via the
// CLI and via the UI.
func TestIntegration_PlanPermission(t *testing.T) {
	t.Parallel()

	// Create org and its owner
	svc := setup(t, nil)
	owner, ownerCtx := svc.createUserCtx(t, ctx)
	org := svc.createOrganization(t, ownerCtx)

	// Create user and add as member of engineers team
	engineer, engineerCtx := svc.createUserCtx(t, ctx)
	team := svc.createTeam(t, ownerCtx, org)
	err := svc.AddTeamMembership(ownerCtx, auth.TeamMembershipOptions{
		TeamID:   team.ID,
		Username: engineer.Username,
	})
	require.NoError(t, err)

	// create some terraform configuration
	configPath := newRootModule(t, svc.Hostname(), org.Name, "my-test-workspace")

	// Open browser and using owner's credentials create a workspace and assign plan
	// role to the engineer's team.
	browser := createBrowserCtx(t)
	err = chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ownerCtx, svc.Hostname(), owner.Username, svc.Secret),
		createWorkspace(t, svc.Hostname(), org.Name, "my-test-workspace"),
		addWorkspacePermission(t, svc.Hostname(), org.Name, "my-test-workspace", team.Name, "plan"),
	})
	require.NoError(t, err)

	// As engineer, run terraform init, and plan. This should succeed because
	// the engineer has been assigned the plan role.
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

	// Now demonstrate UI access by starting a plan via the UI.
	err = chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, svc.Hostname(), engineer.Username, svc.Secret),
		// go to workspace page
		chromedp.Navigate(workspaceURL(svc.Hostname(), org.Name, "my-test-workspace")),
		screenshot(t),
		// select strategy for run
		chromedp.SetValue(`//select[@id="start-run-strategy"]`, "plan-only", chromedp.BySearch),
		screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`body`),
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		screenshot(t),
		chromedp.WaitReady(`#plan-status.phase-status-finished`),
		screenshot(t),
		// wait for run to enter planned-and-finished state
		chromedp.WaitReady(`//*[@class='status status-planned_and_finished']`, chromedp.BySearch),
		screenshot(t),
		// run widget should show plan summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t),
	})
	require.NoError(t, err)
}
