package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_StateUI demonstrates the displaying of terraform state via
// the UI
func TestIntegration_StateUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)
	ws := daemon.createWorkspace(t, ctx, org)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)

	// create run and wait for it to complete
	r := daemon.createRun(t, ctx, ws, cv, nil)
	planned := daemon.waitRunStatus(t, ctx, r.ID, runstatus.Planned)
	err := daemon.Runs.Apply(ctx, planned.ID)
	require.NoError(t, err)
	daemon.waitRunStatus(t, ctx, r.ID, runstatus.Applied)

	sv, err := daemon.State.GetCurrent(ctx, ws.ID)
	require.NoError(t, err)

	// Tests the table of resources and outputs on the workspace overview page.
	t.Run("workspace resources", func(t *testing.T) {
		t.Parallel()

		browser.New(t, ctx, func(page playwright.Page) {
			daemon.gotoPath(t, page, paths.Workspace(ws.ID))

			err = expect.Locator(page.Locator(`//input[@id='resources-label']`)).ToHaveAttribute(`aria-label`, regexp.MustCompile(`Resources \(1\)`))
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`//input[@id='outputs-label']`)).ToHaveAttribute(`aria-label`, regexp.MustCompile(`Outputs \(0\)`))
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[1]`)).ToHaveText(`test`)
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[2]`)).ToHaveText(`hashicorp/null`)
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[3]`)).ToHaveText(`null_resource`)
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[4]`)).ToHaveText(`root`)
			require.NoError(t, err)
		})
	})

	// Tests the state versions list page, verifying versions are shown with
	// serial numbers and navigation links.
	t.Run("list state versions", func(t *testing.T) {
		t.Parallel()

		browser.New(t, ctx, func(page playwright.Page) {
			daemon.gotoPath(t, page, paths.StateVersions(ws.ID))

			// state version row should be present with correct serial
			rowLocator := page.Locator(`#item-state-version-` + sv.ID.String())
			err = expect.Locator(rowLocator).ToBeVisible()
			require.NoError(t, err)

			// the version should be marked current
			err = expect.Locator(rowLocator.Locator(`.badge-success`)).ToHaveText(`current`)
			require.NoError(t, err)

			// serial number should be shown
			err = expect.Locator(rowLocator).ToContainText(fmt.Sprintf("%d", sv.Serial))
			require.NoError(t, err)

			// click View link to navigate to the state version detail page
			err = rowLocator.Locator(`a`, playwright.LocatorLocatorOptions{
				HasText: "View",
			}).Click()
			require.NoError(t, err)

			// detail page should show the raw JSON state file
			err = expect.Locator(page.Locator(`#state-version-json`)).ToContainText(`null_resource`)
			require.NoError(t, err)
		})
	})

	// Tests the state version detail page, verifying the raw JSON contents of
	// the state file are shown.
	t.Run("view state version state json", func(t *testing.T) {
		t.Parallel()

		browser.New(t, ctx, func(page playwright.Page) {
			daemon.gotoPath(t, page, paths.StateVersion(sv.ID))

			// raw JSON should be displayed
			jsonLocator := page.Locator(`#state-version-json`)
			err = expect.Locator(jsonLocator).ToBeVisible()
			require.NoError(t, err)

			// state file should contain terraform_version and resources keys
			err = expect.Locator(jsonLocator).ToContainText(`terraform_version`)
			require.NoError(t, err)

			err = expect.Locator(jsonLocator).ToContainText(`null_resource`)
			require.NoError(t, err)
		})
	})

	// Tests the state version diff page.
	t.Run("diff state versions", func(t *testing.T) {
		t.Parallel()

		browser.New(t, ctx, func(page playwright.Page) {
			daemon.gotoPath(t, page, paths.DiffStateVersion(sv.ID))

			// initial state version: no previous serial, shows "Initial state"
			err = expect.Locator(page.Locator(`body`)).ToContainText(`Initial state`)
			require.NoError(t, err)

			// the resource added should appear in the diff
			err = expect.Locator(page.Locator(`body`)).ToContainText(`null_resource`)
			require.NoError(t, err)
		})
	})
}
