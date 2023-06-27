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
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// Create user and add as member of engineers team
	engineer, engineerCtx := svc.createUserCtx(t)
	team := svc.createTeam(t, ctx, org)
	err := svc.AddTeamMembership(ctx, auth.TeamMembershipOptions{
		TeamID:   team.ID,
		Username: engineer.Username,
	})
	require.NoError(t, err)

	// create some terraform configuration
	configPath := newRootModule(t, svc.Hostname(), org.Name, "my-test-workspace")

	// Open tab and create a workspace and assign plan role to the
	// engineer's team.
	browser.Run(t, ctx, chromedp.Tasks{
		createWorkspace(t, svc.Hostname(), org.Name, "my-test-workspace"),
		addWorkspacePermission(t, svc.Hostname(), org.Name, "my-test-workspace", team.Name, "plan"),
	})

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

	// Now demonstrate engineer can start a plan via the UI.
	browser.Run(t, engineerCtx, chromedp.Tasks{
		// go to workspace page
		chromedp.Navigate(workspaceURL(svc.Hostname(), org.Name, "my-test-workspace")),
		screenshot(t),
		// select operation for run
		chromedp.SetValue(`//select[@id="start-run-operation"]`, "plan-only"),
		screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`),
		screenshot(t),
		chromedp.WaitReady(`#plan-status.phase-status-finished`, chromedp.ByQuery),
		screenshot(t),
		// wait for run to enter planned-and-finished state
		chromedp.WaitReady(`//*[@class='status status-planned_and_finished']`),
		screenshot(t),
		// run widget should show plan summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t),
	})
}
