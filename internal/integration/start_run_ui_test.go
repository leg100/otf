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
	browser.New(t, ctx, startRunTasks(t, svc.System.Hostname(), ws.Organization, ws.Name, run.PlanAndApplyOperation))

	// now destroy resources with the destroy-all operation
	page := browser.New(t, ctx)
		// go to workspace page
		_, err = page.Goto(workspaceURL(svc.System.Hostname(), ws.Organization, ws.Name))
require.NoError(t, err)
		//screenshot(t, "workspace_page"),
		// navigate to workspace settings
		err := page.Locator(`//a[text()='settings']`).Click()
require.NoError(t, err)
		//screenshot(t),
		// click 'queue destroy plan' button
		err := page.Locator(`//button[@id='queue-destroy-plan-button']`).Click()
require.NoError(t, err)
		//screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`),
		//screenshot(t),
		chromedp.WaitReady(`//span[@id='plan-status' and text()='finished']`),
		//screenshot(t),
		// wait for run to enter planned state
		chromedp.WaitReady(`//div[@class='widget']//a[text()='planned']`),
		//screenshot(t),
		// run widget should show plan summary
		matchRegex(t, `//div[@class='widget']//div[@id='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		//screenshot(t),
		// run widget should show discard button
		chromedp.WaitVisible(`//button[@id='run-discard-button']`),
		//screenshot(t),
		// click 'confirm & apply' button once it becomes visible
		err := page.Locator(`//button[text()='apply']`).Click()
require.NoError(t, err)
		//screenshot(t),
		// confirm apply begins and ends
		chromedp.WaitReady(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`),
		chromedp.WaitReady(`//span[@id='apply-status' and text()='finished']`),
		// confirm run ends in applied state
		chromedp.WaitReady(`//div[@class='widget']//a[text()='applied']`),
		// run widget should show apply summary
		matchRegex(t, `//div[@class='widget']//div[@id='resource-summary']`, `\+[0-9]+ \~[0-9]+ \-[0-9]+`),
		//screenshot(t),
	})
}
