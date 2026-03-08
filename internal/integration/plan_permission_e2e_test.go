package integration

import (
	"testing"

	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/user"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_PlanPermission demonstrates the assignment of the workspace
// 'plan' role to a team and what they can and cannot do with that role via the
// CLI and via the UI.
func TestIntegration_PlanPermission(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)

	// Create user and add as member of engineers team
	engineer, engineerCtx := daemon.createUserCtx(t)
	team := daemon.createTeam(t, ctx, org)
	err := daemon.Users.AddTeamMembership(ctx, team.ID, []user.Username{engineer.Username})
	require.NoError(t, err)

	// create some terraform configuration
	configPath := newRootModule(t, daemon.System.Hostname(), org.Name, "my-test-workspace")

	// Open tab and create a workspace and assign plan role to the
	// engineer's team.
	var workspaceURL string
	browser.New(t, ctx, func(page playwright.Page) {
		workspaceURL = createWorkspace(t, page, daemon, org.Name, "my-test-workspace")
		addWorkspacePermission(t, page, workspaceURL, team.ID, "plan")
	})

	// As engineer, run terraform init, and plan. This should succeed because
	// the engineer has been assigned the plan role.
	_ = daemon.engineCLI(t, engineerCtx, "", "init", configPath)
	out := daemon.engineCLI(t, engineerCtx, "", "plan", configPath)
	assert.Contains(t, out, "Plan: 1 to add, 0 to change, 0 to destroy.")

	// Limited privileges should prohibit an apply
	out, err = daemon.engineCLIWithError(t, engineerCtx, "", "apply", configPath, "-auto-approve")
	if assert.Error(t, err) {
		assert.Contains(t, string(out), "Error: Insufficient rights to apply changes")
	}

	// Limited privileges should prohibit a destroy
	out, err = daemon.engineCLIWithError(t, engineerCtx, "", "destroy", configPath, "-auto-approve")
	if assert.Error(t, err) {
		assert.Contains(t, string(out), "Error: Insufficient rights to apply changes")
	}

	// Now demonstrate engineer can start a plan via the UI.
	browser.New(t, ctx, func(page playwright.Page) {
		startRun(t, page, workspaceURL, run.PlanOnlyOperation, false)
	})
}
