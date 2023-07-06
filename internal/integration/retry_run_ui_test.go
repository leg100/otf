package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/run"
)

// TestIntegration_RetryRunUI demonstrates retrying a run via the UI
func TestIntegration_RetryRunUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, nil)
	ws := daemon.createWorkspace(t, ctx, nil)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, &configversion.ConfigurationVersionCreateOptions{
		Speculative: internal.Bool(true),
	})
	// create a run and wait for it reach planned-and-finished state
	r := daemon.createRun(t, ctx, ws, cv)
	for event := range daemon.sub {
		if r, ok := event.Payload.(*run.Run); ok {
			if r.Status == internal.RunErrored {
				t.Fatal("run unexpectedly errored")
			}
			if r.Status == internal.RunPlannedAndFinished {
				break
			}
		}
	}

	// open browser, go to run, and click retry
	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Navigate(runURL(daemon.Hostname(), r.ID)),
		// run should be in planned and finished state
		chromedp.WaitVisible(`//*[@class='status status-planned_and_finished']`),
		screenshot(t, "run_page_planned_and_finished_state"),
		// click retry button
		chromedp.Click(`//button[text()='retry run']`),
		screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`),
		screenshot(t),
		chromedp.WaitVisible(`#plan-status.phase-status-finished`, chromedp.ByQuery),
		// confirm retry button re-appears
		chromedp.WaitVisible(`//button[text()='retry run']`),
		screenshot(t),
	})
}
