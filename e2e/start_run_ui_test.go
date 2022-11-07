package e2e

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

// TestStartRunUI tests starting a run via the Web UI.
func TestStartRunUI(t *testing.T) {
	addBuildsToPath(t)

	user := otf.NewTestUser(t)
	// test using user's personal organization
	org := user.Username()
	hostname := startDaemon(t, user)
	url := "https://" + hostname

	token := createAPIToken(t, hostname)
	login(t, hostname, token)

	workspace := createWebWorkspace(t, allocator, url, org)

	//
	// start run UI requires an existing config version, so create one first
	// by running a plan via the CLI
	//
	root := newRootModule(t, hostname, org, workspace)

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
	require.Contains(t, string(out), "Plan: 1 to add, 0 to change, 0 to destroy.")

	orgSelector := fmt.Sprintf("#item-organization-%s a", org)
	workspaceSelector := fmt.Sprintf("#item-workspace-%s a", workspace)

	allocater := newBrowserAllocater(t)
	ctx, cancel := chromedp.NewContext(allocater)
	defer cancel()

	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		// login
		chromedp.Click(".login-button-github", chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// select org
		chromedp.Click(orgSelector, chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// list workspaces
		chromedp.Click("#menu-item-workspaces > a", chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.WaitReady(`body`),
		// select workspace
		chromedp.Click(workspaceSelector, chromedp.NodeVisible),
		ss.screenshot(t),
		// select strategy for run
		chromedp.SetValue(`//select[@id="start-run-strategy"]`, "plan-and-apply", chromedp.BySearch),
		ss.screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`body`),
		ss.screenshot(t),
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		ss.screenshot(t),
		chromedp.WaitReady(`#plan-status.phase-status-finished`),
		ss.screenshot(t),
		// wait for run to enter planned state
		chromedp.WaitReady(`//*[@id='run-status']//*[normalize-space(text())='planned']`, chromedp.BySearch),
		ss.screenshot(t),
		// click 'confirm & apply' button once it becomes visible
		chromedp.Click(`//button[text()='Confirm & Apply']`, chromedp.NodeVisible, chromedp.BySearch),
		ss.screenshot(t),
		// confirm apply begins and ends
		chromedp.WaitReady(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		chromedp.WaitReady(`#apply-status.phase-status-finished`),
		// confirm run ends in applied state
		chromedp.WaitReady(`//*[@id='run-status']//*[normalize-space(text())='applied']`, chromedp.BySearch),
	})
	require.NoError(t, err)
}
