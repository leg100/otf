package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	expect "github.com/google/goexpect"
	"github.com/stretchr/testify/require"
)

func githubLoginTasks(t *testing.T, hostname string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("https://" + hostname),
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
	}
}

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
		chromedp.ActionFunc(func(ctx context.Context) error {
			var got string
			err := chromedp.Run(ctx, chromedp.Text(".flash-success", &got, chromedp.NodeVisible))
			if err != nil {
				return err
			}
			require.Equal(t, "created workspace: "+name, strings.TrimSpace(got))
			return nil
		}),
	}
}

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
		chromedp.Navigate("https://" + hostname),
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		chromedp.Click("#top-right-profile-link > a", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		chromedp.Click("#user-tokens-link > a", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		chromedp.Click("#new-user-token-button", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		chromedp.Focus("#description", chromedp.NodeVisible),
		input.InsertText("e2e-test"),
		chromedp.Submit("#description"),
		chromedp.WaitReady(`body`),
		chromedp.Text(".flash-success > .data", &token, chromedp.NodeVisible),
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
