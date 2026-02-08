package integration

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestAutoApply demonstrates enabling auto-apply via the UI.
func TestAutoApply(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

	// create workspace and enable auto-apply
	browser.New(t, ctx, func(page playwright.Page) {
		createWorkspace(t, page, svc.System.Hostname(), org.Name, t.Name())
		// go to workspace
		_, err := page.Goto(workspaceURL(svc.System.Hostname(), org.Name, t.Name()))
		require.NoError(t, err)
		// go to workspace settings
		err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
		require.NoError(t, err)
		// enable auto-apply
		err = page.Locator(`//input[@name='auto_apply' and @value='true']`).Click()
		require.NoError(t, err)
		// submit form
		err = page.Locator(`//button[text()='Save changes']`).Click()
		require.NoError(t, err)
		// confirm workspace updated
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated workspace")
		require.NoError(t, err)
		// check UI has correctly updated the workspace resource
		ws, err := svc.Workspaces.GetByName(ctx, org.Name, t.Name())
		require.NoError(t, err)
		require.Equal(t, true, ws.AutoApply)
	})
}
