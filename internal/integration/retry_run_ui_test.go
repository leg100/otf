package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RetryRunUI demonstrates retrying a run via the UI
func TestIntegration_RetryRunUI(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	user, ctx := daemon.createUserCtx(t, ctx)
	ws := daemon.createWorkspace(t, ctx, nil)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, &configversion.ConfigurationVersionCreateOptions{
		Speculative: internal.Bool(true),
	})
	sub := daemon.createSubscriber(t, ctx)

	// create a run and wait for it reach planned-and-finished state
	r := daemon.createRun(t, ctx, ws, cv)
	for event := range sub {
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
	browser := createBrowserCtx(t)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, daemon.Hostname(), user.Username, daemon.Secret),
		chromedp.Navigate(runURL(daemon.Hostname(), r.ID)),
		// run should be in planned and finished state
		chromedp.WaitReady(`//*[@class='status status-planned_and_finished']`, chromedp.BySearch),
		screenshot(t, "run_page_planned_and_finished_state"),
		// click retry button
		chromedp.Click(`//button[text()='retry run']`, chromedp.NodeVisible, chromedp.BySearch),
		screenshot(t),
		// confirm plan begins and ends
		chromedp.WaitReady(`body`),
		chromedp.WaitReady(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`, chromedp.BySearch),
		screenshot(t),
		chromedp.WaitReady(`#plan-status.phase-status-finished`),
		// confirm retry button re-appears
		chromedp.WaitReady(`//button[text()='retry run']`, chromedp.BySearch),
		screenshot(t),
	})
	require.NoError(t, err)
}
