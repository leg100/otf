package integration

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunListUI demonstrates listing runs via the UI.
func TestIntegration_RunListUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, nil)
	ws := daemon.createWorkspace(t, ctx, nil)
	tfConfig := newRootModule(t, daemon.System.Hostname(), ws.Organization, ws.Name)

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

	// meanwhile, execute a terraform cli init and plan
	daemon.tfcli(t, ctx, "init", tfConfig)
	daemon.tfcli(t, ctx, "plan", tfConfig)

	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to runs page
		_, err := page.Goto(runsURL(daemon.System.Hostname(), ws.ID))
		require.NoError(t, err)

		// should be one run listed with status planned and finished
		err = expect.Locator(page.GetByText(`planned and finished`)).ToHaveCount(1)
		require.NoError(t, err)
	})
}
