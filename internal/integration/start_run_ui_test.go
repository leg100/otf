package integration

import (
	"regexp"
	"testing"

	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
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
	page := browser.New(t, ctx)

	startRunTasks(t, page, svc.System.Hostname(), ws.Organization, ws.Name, run.PlanAndApplyOperation)

	// now destroy resources with the destroy-all operation
	// go to workspace page
	_, err := page.Goto(workspaceURL(svc.System.Hostname(), ws.Organization, ws.Name))
	require.NoError(t, err)
	//screenshot(t, "workspace_page"),
	// navigate to workspace settings
	err = page.Locator(`//a[text()='settings']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// click 'queue destroy plan' button
	err = page.Locator(`//button[@id='queue-destroy-plan-button']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// confirm plan begins and ends
	err = expect.Locator(page.Locator(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`)).ToBeVisible()
	require.NoError(t, err)
	//screenshot(t),

	err = expect.Locator(page.Locator(`//span[@id='plan-status' and text()='finished']`)).ToBeAttached()
	require.NoError(t, err)
	//screenshot(t),

	// wait for run to enter planned state
	err = expect.Locator(page.Locator(`//div[@class='widget']//a[text()='planned']`)).ToBeAttached()
	require.NoError(t, err)
	//screenshot(t),

	// run widget should show plan summary
	err = expect.Locator(page.Locator(`//div[@class='widget']//div[@id='resource-summary']`)).ToHaveText(regexp.MustCompile(`\+[0-9]+ \~[0-9]+ \-[0-9]+`))
	require.NoError(t, err)
	//screenshot(t),

	// run widget should show discard button
	err = expect.Locator(page.Locator(`//button[@id='run-discard-button']`)).ToBeVisible()
	require.NoError(t, err)
	//screenshot(t),

	// click 'confirm & apply' button once it becomes visible
	err = page.Locator(`//button[text()='apply']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// confirm apply begins and ends
	err = expect.Locator(page.Locator(`//*[@id='tailed-apply-logs']//text()[contains(.,'Initializing the backend')]`)).ToBeAttached()
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//span[@id='apply-status' and text()='finished']`)).ToBeAttached()
	require.NoError(t, err)

	// confirm run ends in applied state
	err = expect.Locator(page.Locator(`//div[@class='widget']//a[text()='applied']`)).ToBeAttached()
	require.NoError(t, err)

	// run widget should show apply summary
	err = expect.Locator(page.Locator(`//div[@class='widget']//div[@id='resource-summary']`)).ToHaveText(regexp.MustCompile(`\+[0-9]+ \~[0-9]+ \-[0-9]+`))
	require.NoError(t, err)
	//screenshot(t),
}
