package integration

import (
	"testing"

	"github.com/leg100/otf/internal/user"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestWritePermissionE2E demonstrates a user with write permissions on a workspace interacting
// with the workspace via the terraform CLI.
func TestWritePermissionE2E(t *testing.T) {
	integrationTest(t)

	// Create user and org, and user becomes owner of the org
	daemon, org, ctx := setup(t)

	// Create engineer user and team and make member of a team
	engineer, engineerCtx := daemon.createUserCtx(t)
	team := daemon.createTeam(t, ctx, org)
	err := daemon.Users.AddTeamMembership(ctx, team.ID, []user.Username{engineer.Username})
	require.NoError(t, err)

	// create terraform config
	config := newRootModule(t, daemon.System.Hostname(), org.Name, "my-test-workspace")

	// Open browser, create workspace and assign write permissions to the
	// engineer's team.
	browser.New(t, ctx, func(page playwright.Page) {
		workspaceURL := createWorkspace(t, page, daemon, org.Name, "my-test-workspace")
		addWorkspacePermission(t, page, workspaceURL, team.ID, "write")
	})

	// As engineer, run terraform init
	_ = daemon.engineCLI(t, engineerCtx, "", "init", config)

	out := daemon.engineCLI(t, engineerCtx, "", "plan", config)
	require.Contains(t, out, "Plan: 1 to add, 0 to change, 0 to destroy.")

	out = daemon.engineCLI(t, engineerCtx, "", "apply", config, "-auto-approve")
	require.Contains(t, out, "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")

	out = daemon.engineCLI(t, engineerCtx, "", "destroy", config, "-auto-approve")
	require.Contains(t, out, "Apply complete! Resources: 0 added, 0 changed, 1 destroyed.")

	// lock and unlock workspace
	daemon.otfCLI(t, ctx, "workspaces", "lock", "my-test-workspace", "--organization", org.Name.String())
	daemon.otfCLI(t, ctx, "workspaces", "unlock", "my-test-workspace", "--organization", org.Name.String())

	// list workspaces
	out = daemon.otfCLI(t, ctx, "workspaces", "list", "--organization", org.Name.String())
	require.Contains(t, out, "my-test-workspace")
}
