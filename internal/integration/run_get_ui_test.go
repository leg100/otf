package integration

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunGetUI tests the main run web page.
func TestIntegration_RunGetUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)
	ws1 := daemon.createWorkspace(t, ctx, org)
	cv1 := daemon.createAndUploadConfigurationVersion(t, ctx, ws1, nil)
	run1 := daemon.createRun(t, ctx, ws1, cv1, nil)

	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to run page
		_, err := page.Goto(runURL(daemon.System.Hostname(), run1.ID))
		require.NoError(t, err)

		// click clipboard icon to copy run ID into clipboard
		err = page.Locator(`//div[@id='run-identifier']//img[@id='clipboard-icon']`).Click()
		require.NoError(t, err)

		// read run ID from clipboard and check it matches actual run ID
		clipboardContents, err := page.EvaluateHandle(`window.navigator.clipboard.readText()`)
		require.NoError(t, err)
		assert.Equal(t, run1.ID.String(), clipboardContents.String())
	})
}
