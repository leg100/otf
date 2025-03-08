package integration

import (
	"testing"

	"github.com/leg100/otf/internal/runstatus"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunListUI demonstrates listing runs via the UI.
func TestIntegration_RunListUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, nil)
	ws := daemon.createWorkspace(t, ctx, nil)

	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to workspace page
		_, err := page.Goto(workspaceURL(daemon.System.Hostname(), ws.Organization, ws.Name))
		require.NoError(t, err)

		// navigate to runs page
		err = page.Locator(`//a[text()='runs']`).Click()
		require.NoError(t, err)

		// should be no runs listed
		err = expect.Locator(page.Locator(`//div[@id='content-list']`)).ToHaveText(`No items currently exist.`)
		require.NoError(t, err)
	})

	// create several runs
	cv1 := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	cv2 := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	cv3 := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)
	// create run, and apply
	run1 := daemon.createRun(t, ctx, ws, cv1, nil)
	{
		_ = daemon.waitRunStatus(t, run1.ID, runstatus.Planned)
		err := daemon.Runs.Apply(ctx, run1.ID)
		require.NoError(t, err)
		_ = daemon.waitRunStatus(t, run1.ID, runstatus.Applied)
	}
	// create two runs, which should reached planned&finished state.
	_ = daemon.createRun(t, ctx, ws, cv2, nil)
	_ = daemon.createRun(t, ctx, ws, cv3, nil)

	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to runs page
		_, err := page.Goto(runsURL(daemon.System.Hostname(), ws.ID))
		require.NoError(t, err)

		// should be three runs
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-3 of 3")
		require.NoError(t, err)

		// show status filter
		err = page.Locator(`#toggle-status-filter-visibility`).Click()
		require.NoError(t, err)

		// filter by planned&finished
		err = page.Locator(`#filter-status-planned_and_finished`).Click()
		require.NoError(t, err)

		// should only show two runs
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-2 of 2")
		require.NoError(t, err)
	})
}
