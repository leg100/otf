package integration

import (
	"testing"

	"github.com/leg100/otf/internal/ui/paths"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestRunTriggersUI tests run trigger UI functionality.
func TestRunTriggersUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)
	ws1 := daemon.createWorkspace(t, ctx, org)
	ws2 := daemon.createWorkspace(t, ctx, org)

	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to workspace page
		_, err := page.Goto(daemon.URL(paths.Workspace(ws1.ID)))
		require.NoError(t, err)

		// navigate to settings page
		err = page.Locator("#menu-item-settings > a").Click()
		require.NoError(t, err)

		// navigate to run triggers page
		err = page.Locator("#menu-item-run-triggers > a").Click()
		require.NoError(t, err)

		// connect to ws2
		_, err = page.Locator(`//select[@id="connect_workspace"]`).SelectOption(playwright.SelectOptionValues{
			ValuesOrLabels: new([]string{ws2.ID.String()}),
		})
		require.NoError(t, err)

		// form should auto submit and redirect to same page with ws2 listed as
		// connected workspace.
		err = expect.Locator(page.Locator(`//table/tbody/tr/td[1]`)).ToHaveText(ws2.Name)
		require.NoError(t, err)

		// now delete the connection to ws2
		err = page.Locator(`//table/tbody/tr/td[2]//button`).Click()
		require.NoError(t, err)

		// should redirect to same page and show alert.
		err = expect.Locator(page.GetByRole("alert")).ToContainText("deleted trigger")
		require.NoError(t, err)
	})
}
