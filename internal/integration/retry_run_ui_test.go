package integration

import (
	"testing"

	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RetryRunUI demonstrates retrying a run via the UI
func TestIntegration_RetryRunUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t)
	ws := daemon.createWorkspace(t, ctx, nil)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, &configversion.CreateOptions{
		Speculative: new(true),
	})
	// create a run and wait for it reach planned-and-finished state
	r := daemon.createRun(t, ctx, ws, cv, nil)
	daemon.waitRunStatus(t, ctx, r.ID, runstatus.PlannedAndFinished)

	// open browser, go to run, and click retry
	browser.New(t, ctx, func(page playwright.Page) {
		_, err := page.Goto(runURL(daemon.System.Hostname(), r.ID))
		require.NoError(t, err)
		// run should be in planned and finished state
		err = expect.Locator(page.Locator(`//a[text()='planned and finished']`)).ToBeVisible()
		require.NoError(t, err)
		screenshot(t, page, "run_page_planned_and_finished_state")
		// click retry button
		err = page.Locator(`//button[@id='retry-button']`).Click()
		require.NoError(t, err)
		// confirm plan begins and ends
		expect.Locator(page.Locator(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`))

		err = expect.Locator(page.Locator(`//span[@id='plan-status' and text()='finished']`)).ToBeVisible()
		require.NoError(t, err)

		// confirm retry button re-appears
		err = expect.Locator(page.Locator(`//button[@id='retry-button']`)).ToBeVisible()
		require.NoError(t, err)
	})
}
