package e2e

import (
	"context"
	"os/exec"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlanPermission demonstrates a user with plan permissions on a workspace interacting
// with the workspace via the terraform CLI.
func TestPlanPermission(t *testing.T) {
	addBuildsToPath(t)

	workspaceName := "plan-perms"

	// First we need to setup an organization with a user who is both in the
	// owners team and the devops team.
	org := uuid.NewString()
	owners := cloud.Team{Name: "owners", Organization: org}
	devops := cloud.Team{Name: "devops", Organization: org}

	// Build and start a daemon specifically for the boss
	boss := cloud.User{
		Name:          "boss-" + uuid.NewString(),
		Organizations: []string{org},
		Teams: []cloud.Team{
			owners,
			devops,
		},
	}
	bossDaemon := &daemon{}
	bossDaemon.withGithubUser(&boss)
	bossHostname := bossDaemon.start(t)

	// setup non-owner engineer user - note we start another daemon because this is the
	// only way at present that an additional user can be seeded for testing.
	engineer := cloud.User{
		Name:          "engineer-" + uuid.NewString(),
		Organizations: []string{org},
		Teams: []cloud.Team{
			devops,
		},
	}
	engineerDaemon := &daemon{}
	engineerDaemon.withGithubUser(&engineer)
	engineerHostname := engineerDaemon.start(t)

	// create terraform configPath
	configPath := newRootModule(t, engineerHostname, org, workspaceName)

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Tasks{
		// login to UI as boss
		githubLoginTasks(t, bossHostname, boss.Name),
		// create workspace via UI
		createWorkspaceTasks(t, bossHostname, org, workspaceName),
		// assign plan permissions to devops team
		addWorkspacePermissionTasks(t, bossHostname, org, workspaceName, devops.Name, "plan"),
		// logout of UI (as boss)
		logoutTasks(t, bossHostname),
		// login to UI as engineer
		githubLoginTasks(t, engineerHostname, engineer.Name),
		// create api token and run terraform login (as engineer)
		terraformLoginTasks(t, engineerHostname),
		// terraform init (as engineer)
		terraformInitTasks(t, configPath),
		// terraform plan (as engineer)
		chromedp.ActionFunc(func(ctx context.Context) error {
			cmd := exec.Command("terraform", "plan", "-no-color")
			cmd.Dir = configPath
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			require.NoError(t, err)
			assert.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")
			return nil
		}),
		// terraform apply (as engineer)
		chromedp.ActionFunc(func(ctx context.Context) error {
			cmd := exec.Command("terraform", "apply", "-no-color", "-auto-approve")
			cmd.Dir = configPath
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			if assert.Error(t, err) {
				assert.Contains(t, string(out), "Error: Insufficient rights to apply changes")
			}
			return nil
		}),
		// terraform destroy (as engineer)
		chromedp.ActionFunc(func(ctx context.Context) error {
			cmd := exec.Command("terraform", "destroy", "-no-color", "-auto-approve")
			cmd.Dir = configPath
			out, err := cmd.CombinedOutput()
			t.Log(string(out))
			if assert.Error(t, err) {
				assert.Contains(t, string(out), "Error: Insufficient rights to apply changes")
			}
			return nil
		}),
	})
	require.NoError(t, err)
}
