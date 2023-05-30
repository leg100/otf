package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// TestStartRunUI tests starting a run via the Web UI before confirming and
// applying the run.
func TestStartRunUI(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)

	user, ctx := svc.createUserCtx(t, ctx)
	ws := svc.createWorkspace(t, ctx, nil)
	_ = svc.createAndUploadConfigurationVersion(t, ctx, ws, nil)

	// now we have a config version, start a run with the plan-and-apply
	// strategy
	browser := createBrowserCtx(t)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, svc.Hostname(), user.Username, svc.Secret),
		startRunTasks(t, svc.Hostname(), ws.Organization, ws.Name, "plan-and-apply"),
	})
	require.NoError(t, err)

	// now destroy resources with the destroy-all strategy
	okDialog(t, browser)
	err = chromedp.Run(browser, chromedp.Tasks{
		// go to workspace page
		chromedp.Navigate(workspaceURL(svc.Hostname(), ws.Organization, ws.Name)),
		screenshot(t, "workspace_page"),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// click 'queue destroy plan' button
		chromedp.Click(`//button[@id='queue-destroy-plan-button']`, chromedp.BySearch),
		screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`body`),
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		screenshot(t),
		chromedp.WaitReady(`#plan-status.phase-status-finished`),
		screenshot(t),
		// wait for run to enter planned state
		chromedp.WaitReady(`//*[@class='status status-planned']`, chromedp.BySearch),
		screenshot(t),
		// run widget should show plan summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t),
		// run widget should show discard button
		chromedp.WaitReady(`//button[@id='run-discard-button']`, chromedp.BySearch),
		screenshot(t),
		// click 'confirm & apply' button once it becomes visible
		chromedp.Click(`//button[text()='apply']`, chromedp.NodeVisible, chromedp.BySearch),
		screenshot(t),
		// confirm apply begins and ends
		chromedp.WaitReady(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		chromedp.WaitReady(`#apply-status.phase-status-finished`),
		// confirm run ends in applied state
		chromedp.WaitReady(`//*[@class='status status-applied']`, chromedp.BySearch),
		// run widget should show apply summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t),
	})
	require.NoError(t, err)
}
