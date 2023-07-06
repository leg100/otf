package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/run"
)

// TestStartRunUI tests starting a run via the Web UI before confirming and
// applying the run.
func TestStartRunUI(t *testing.T) {
	integrationTest(t)

	svc, _, ctx := setup(t, nil)

	ws := svc.createWorkspace(t, ctx, nil)
	_ = svc.createAndUploadConfigurationVersion(t, ctx, ws, nil)

	// now we have a config version, start a run with the plan-and-apply
	// operation
	browser.Run(t, ctx, startRunTasks(t, svc.Hostname(), ws.Organization, ws.Name, run.PlanAndApplyOperation))

	// now destroy resources with the destroy-all operation
	browser.Run(t, ctx, chromedp.Tasks{
		// go to workspace page
		chromedp.Navigate(workspaceURL(svc.Hostname(), ws.Organization, ws.Name)),
		screenshot(t, "workspace_page"),
		// navigate to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		screenshot(t),
		// click 'queue destroy plan' button
		chromedp.Click(`//button[@id='queue-destroy-plan-button']`),
		screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`),
		screenshot(t),
		chromedp.WaitReady(`#plan-status.phase-status-finished`, chromedp.ByQuery),
		screenshot(t),
		// wait for run to enter planned state
		chromedp.WaitReady(`//*[@class='status status-planned']`),
		screenshot(t),
		// run widget should show plan summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t),
		// run widget should show discard button
		chromedp.WaitVisible(`//button[@id='run-discard-button']`),
		screenshot(t),
		// click 'confirm & apply' button once it becomes visible
		chromedp.Click(`//button[text()='apply']`),
		screenshot(t),
		// confirm apply begins and ends
		chromedp.WaitReady(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`),
		chromedp.WaitReady(`#apply-status.phase-status-finished`, chromedp.ByQuery),
		// confirm run ends in applied state
		chromedp.WaitReady(`//*[@class='status status-applied']`),
		// run widget should show apply summary
		matchRegex(t, `//div[@class='item']//div[@class='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		screenshot(t),
	})
}
