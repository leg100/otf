package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIntegration_RunListUI demonstrates listing runs via the UI.
func TestIntegration_RunListUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, nil)
	ws := daemon.createWorkspace(t, ctx, nil)
	tfConfig := newRootModule(t, daemon.System.Hostname(), ws.Organization, ws.Name)

	page := browser.New(t, ctx)

	// navigate to workspace page
	_, err := page.Goto(workspaceURL(daemon.System.Hostname(), ws.Organization, ws.Name))
	require.NoError(t, err)

	// navigate to runs page
	err = page.Locator(`//a[text()='runs']`).Click()
	require.NoError(t, err)

	// should be no runs listed
	err = expect.Locator(page.Locator(`//div[@id='content-list']`)).ToHaveText(`No items currently exist.`)
	require.NoError(t, err)

	// meanwhile, execute a terraform cli init and plan
	daemon.tfcli(t, ctx, "init", tfConfig)
	daemon.tfcli(t, ctx, "plan", tfConfig)

	// should be one run listed
	err = expect.Locator(page.Locator(`//div[@id='content-list']//*[@class='widget']`)).ToHaveCount(1)
	require.NoError(t, err)

	// and its status should be 'planned and finished'
	err = expect.Locator(page.Locator(`//*[@class='widget']//a[text()='planned and finished']`)).ToBeVisible()
	require.NoError(t, err)
}
