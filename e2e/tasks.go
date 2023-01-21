package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	expect "github.com/google/goexpect"
	"github.com/stretchr/testify/require"
)

// githubLoginTasks logs into the UI using github; upon success a session cookie
// is created
func githubLoginTasks(t *testing.T, hostname, username string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to login page
		chromedp.Navigate("https://" + hostname + "/login"),
		chromedp.WaitReady(`body`),
		// login
		chromedp.Click("a.login-button-github", chromedp.NodeVisible),
		screenshot(t),
		// check login confirmation message
		matchText(t, ".content > p", "You are logged in as "+username),
	}
}

// logoutTasks logs out of the UI
func logoutTasks(t *testing.T, hostname string) chromedp.Tasks {
	var gotLoginLocation string
	return chromedp.Tasks{
		// go to profile
		chromedp.Click("#top-right-profile-link > a", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// logout
		chromedp.Click("button#logout", chromedp.NodeVisible),
		screenshot(t),
		// should be redirected to login page
		chromedp.Location(&gotLoginLocation),
		chromedp.ActionFunc(func(ctx context.Context) error {
			require.Equal(t, fmt.Sprintf("https://%s/login", hostname), gotLoginLocation)
			return nil
		}),
	}
}

// createWorkspaceTasks creates a workspace via the UI
func createWorkspaceTasks(t *testing.T, hostname, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("https://" + hostname + "/organizations/" + org),
		screenshot(t),
		chromedp.Click("#menu-item-workspaces > a", chromedp.ByQuery),
		screenshot(t),
		chromedp.Click("#new-workspace-button", chromedp.NodeVisible, chromedp.ByQuery),
		screenshot(t),
		chromedp.Focus("input#name", chromedp.NodeVisible),
		input.InsertText(name),
		chromedp.Click("#create-workspace-button"),
		screenshot(t),
		matchText(t, ".flash-success", "created workspace: "+name),
	}
}

// startRunTasks starts a run via the UI
func startRunTasks(t *testing.T, hostname, organization string, workspaceName string) chromedp.Tasks {
	return []chromedp.Action{
		// go to workspace page
		chromedp.Navigate(fmt.Sprintf("https://%s/organizations/%s/workspaces/%s", hostname, organization, workspaceName)),
		screenshot(t),
		// select strategy for run
		chromedp.SetValue(`//select[@id="start-run-strategy"]`, "plan-and-apply", chromedp.BySearch),
		screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`body`),
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		screenshot(t),
		chromedp.WaitReady(`#plan-status.phase-status-finished`),
		screenshot(t),
		// wait for run to enter planned state
		chromedp.WaitReady(`//*[@id='run-status']//*[normalize-space(text())='planned']`, chromedp.BySearch),
		screenshot(t),
		// click 'confirm & apply' button once it becomes visible
		chromedp.Click(`//button[text()='Confirm & Apply']`, chromedp.NodeVisible, chromedp.BySearch),
		screenshot(t),
		// confirm apply begins and ends
		chromedp.WaitReady(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		chromedp.WaitReady(`#apply-status.phase-status-finished`),
		// confirm run ends in applied state
		chromedp.WaitReady(`//*[@id='run-status']//*[normalize-space(text())='applied']`, chromedp.BySearch),
		screenshot(t),
	}
}

// terraformLoginTasks creates an API token via the UI before passing it to
// 'terraform login'
func terraformLoginTasks(t *testing.T, hostname string) chromedp.Tasks {
	var token string
	return []chromedp.Action{
		// go to profile
		chromedp.Click("#top-right-profile-link > a", chromedp.NodeVisible),
		screenshot(t),
		// go to tokens
		chromedp.Click("#user-tokens-link > a", chromedp.NodeVisible),
		screenshot(t),
		// create new token
		chromedp.Click("#new-user-token-button", chromedp.NodeVisible),
		screenshot(t),
		chromedp.Focus("#description", chromedp.NodeVisible),
		input.InsertText("e2e-test"),
		chromedp.Submit("#description"),
		screenshot(t),
		// capture token
		chromedp.Text(".flash-success > .data", &token, chromedp.NodeVisible),
		// pass token to terraform login
		chromedp.ActionFunc(func(ctx context.Context) error {
			tfpath, err := exec.LookPath("terraform")
			require.NoErrorf(t, err, "terraform executable not found in path")

			// nullifying PATH makes `terraform login` skip opening a browser
			// window
			path := os.Getenv("PATH")
			os.Setenv("PATH", "")
			defer os.Setenv("PATH", path)

			e, tferr, err := expect.SpawnWithArgs(
				[]string{tfpath, "login", hostname},
				time.Minute,
				expect.PartialMatch(true),
				expect.Verbose(testing.Verbose()))
			require.NoError(t, err)
			defer e.Close()

			e.ExpectBatch([]expect.Batcher{
				&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: "yes\n"},
				&expect.BExp{R: "Enter a value:"}, &expect.BSnd{S: token + "\n"},
				&expect.BExp{R: "Success! Logged in to Terraform Enterprise"},
			}, time.Minute)
			return <-tferr
		}),
	}
}

// addWorkspacePermissionTasks adds a workspace permission via the UI, assigning
// a role to a team.
func addWorkspacePermissionTasks(t *testing.T, url, org, workspaceName, team, role string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(path.Join(url, "organizations", org, "workspaces", workspaceName)),
		screenshot(t),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// confirm builtin admin permission for owners team
		matchText(t, "#permissions-owners td:first-child", "owners"),
		matchText(t, "#permissions-owners td:last-child", "admin"),
		// assign role to team
		chromedp.SetValue(`//select[@id="permissions-add-select-role"]`, role, chromedp.BySearch),
		chromedp.SetValue(`//select[@id="permissions-add-select-team"]`, team, chromedp.BySearch),
		chromedp.Click("#permissions-add-button", chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "updated workspace permissions"),
	}
}

func terraformInitTasks(t *testing.T, path string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		cmd := exec.Command("terraform", "init", "-no-color")
		cmd.Dir = path
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		return err
	})
}

func terraformPlanTasks(t *testing.T, root string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		cmd := exec.Command("terraform", "plan", "-no-color")
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		t.Log(string(out))
		require.NoError(t, err)
		require.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")
		return nil
	})
}

func createGithubVCSProviderTasks(t *testing.T, url, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to org
		chromedp.Navigate(path.Join(url, "organizations", org)),
		// go to vcs providers
		chromedp.Click("#vcs_providers > a", chromedp.NodeVisible),
		screenshot(t),
		// click 'New Github VCS Provider' button
		chromedp.Click(`//button[text()='New Github VCS Provider']`, chromedp.NodeVisible),
		screenshot(t),
		// enter fake github token and name
		chromedp.Focus("input#token", chromedp.NodeVisible),
		input.InsertText("fake-github-personal-token"),
		chromedp.Focus("input#name"),
		input.InsertText(name),
		screenshot(t),
		// submit form to create provider
		chromedp.Submit("input#token"),
		screenshot(t),
		matchText(t, ".flash-success", "created provider: github"),
	}
}

func connectWorkspaceTasks(t *testing.T, url, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(path.Join(url, "organizations", org, "workspaces", name)),
		screenshot(t),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// click connect button
		chromedp.Click(`//button[text()='Connect to VCS']`, chromedp.NodeVisible),
		screenshot(t),
		// select provider
		chromedp.Click(`//a[normalize-space(text())='github']`, chromedp.NodeVisible),
		screenshot(t),
		// connect to first repo in list (there should only be one)
		chromedp.Click(`//div[@class='content-list']//button[text()='connect']`, chromedp.NodeVisible),
		screenshot(t),
		// confirm connected
		matchText(t, ".flash-success", "connected workspace to repo"),
	}
}

func disconnectWorkspaceTasks(t *testing.T, url, org, name string) chromedp.Tasks {
	return chromedp.Tasks{
		// go to workspace
		chromedp.Navigate(path.Join(url, "organizations", org, "workspaces", name)),
		screenshot(t),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// click disconnect button
		chromedp.Click(`//button[@id='disconnect-workspace-repo-button']`, chromedp.NodeVisible),
		screenshot(t),
		// confirm disconnected
		matchText(t, ".flash-success", "disconnected workspace from repo"),
	}
}
