package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RetryRunUI demonstrates retrying a run via the UI
func TestIntegration_RetryRunUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, nil)
	ws := daemon.createWorkspace(t, ctx, nil)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, &configversion.CreateOptions{
		Speculative: internal.Bool(true),
	})
	// watch run events
	sub, unsub := daemon.Runs.Watch(ctx)
	defer unsub()
	// create a run and wait for it reach planned-and-finished state
	r := daemon.createRun(t, ctx, ws, cv)
	for event := range sub {
		if event.Payload.Status == run.RunErrored {
			t.Fatal("run unexpectedly errored")
		}
		if event.Payload.Status == run.RunPlannedAndFinished {
			break
		}
	}

	// open browser, go to run, and click retry
	page := browser.New(t, ctx)
	_, err := page.Goto(runURL(daemon.System.Hostname(), r.ID))
	require.NoError(t, err)
	// run should be in planned and finished state
	err = expect.Locator(page.Locator(`//a[text()='planned and finished']`)).ToBeVisible()
	require.NoError(t, err)
	screenshot(t, page, "run_page_planned_and_finished_state")
	// click retry button
	err = page.Locator(`//button[text()='retry run']`).Click()
	require.NoError(t, err)
	// confirm plan begins and ends
	expect.Locator(page.Locator(`//*[@id='tailed-plan-logs']//text()[contains(.,'Initializing the backend')]`))

	err = expect.Locator(page.Locator(`//span[@id='plan-status' and text()='finished']`)).ToBeVisible()
	require.NoError(t, err)

	// confirm retry button re-appears
	err = expect.Locator(page.Locator(`//button[text()='retry run']`)).ToBeVisible()
	require.NoError(t, err)
}
