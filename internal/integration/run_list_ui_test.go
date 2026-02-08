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

	daemon, org, ctx := setup(t)
	ws1 := daemon.createWorkspace(t, ctx, org)
	ws2 := daemon.createWorkspace(t, ctx, org)

	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to workspace page
		_, err := page.Goto(workspaceURL(daemon.System.Hostname(), ws1.Organization, ws1.Name))
		require.NoError(t, err)

		// navigate to runs page
		err = page.Locator(`//li[@id='menu-item-runs']/a`).Click()
		require.NoError(t, err)

		// should be no runs listed
		err = expect.Locator(page.Locator(`//*[@id='no-items-found']`)).ToHaveText(`No items found`)
		require.NoError(t, err)
	})

	// create several runs
	cv1 := daemon.createAndUploadConfigurationVersion(t, ctx, ws1, nil)
	cv2 := daemon.createAndUploadConfigurationVersion(t, ctx, ws1, nil)
	cv3 := daemon.createAndUploadConfigurationVersion(t, ctx, ws1, nil)
	cv4 := daemon.createAndUploadConfigurationVersion(t, ctx, ws2, nil)
	// create run, and apply
	run1 := daemon.createRun(t, ctx, ws1, cv1, nil)
	{
		_ = daemon.waitRunStatus(t, ctx, run1.ID, runstatus.Planned)
		err := daemon.Runs.Apply(ctx, run1.ID)
		require.NoError(t, err)
		_ = daemon.waitRunStatus(t, ctx, run1.ID, runstatus.Applied)
	}
	// create two runs, which should reached planned&finished state.
	_ = daemon.createRun(t, ctx, ws1, cv2, nil)
	_ = daemon.createRun(t, ctx, ws1, cv3, nil)
	// create another run on a different workspace
	_ = daemon.createRun(t, ctx, ws2, cv4, nil)

	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to workspace runs page
		_, err := page.Goto(workspaceRunsURL(daemon.System.Hostname(), ws1.ID))
		require.NoError(t, err)

		// confirm 'runs' submenu button is active
		err = expect.Locator(page.Locator(`//li[@id='menu-item-runs']/a`)).ToHaveClass(`menu-active`)
		require.NoError(t, err)

		// should be three runs
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-3 of 3")
		require.NoError(t, err)

		// show status filter
		err = page.Locator(`//input[@name='status_filter_visible']`).Click()
		require.NoError(t, err)

		// filter by planned&finished
		//
		// NOTE: Evaluate() is used in place of Click() because the latter is
		// notoriously flaky when used with customized checkboxes:
		//
		// https://github.com/microsoft/playwright/issues/13470
		_, err = page.Locator(`//input[@id='filter-item-planned_and_finished']`).Evaluate(`el => el.click()`, nil)
		require.NoError(t, err)

		// should only show two runs
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-2 of 2")
		require.NoError(t, err)
	})

	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to organization runs page
		_, err := page.Goto(organizationRunsURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)

		// should be four runs
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-4 of 4")
		require.NoError(t, err)

		// show status filter
		err = page.Locator(`//input[@name='status_filter_visible']`).Click()
		require.NoError(t, err)

		// filter by planned&finished
		//
		// NOTE: Evaluate() is used in place of Click() because the latter is
		// notoriously flaky when used with customized checkboxes:
		//
		// https://github.com/microsoft/playwright/issues/13470
		_, err = page.Locator(`//input[@id='filter-item-planned_and_finished']`).Evaluate(`el => el.click()`, nil)
		require.NoError(t, err)

		// should only show two runs
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-2 of 2")
		require.NoError(t, err)
	})
}
