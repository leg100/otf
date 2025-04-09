package integration

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestRunnersUI demonstrates managing runners via the UI
func TestRunnersUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	browser.New(t, ctx, func(page playwright.Page) {
		// go to org main menu
		_, err := page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)

		// to list of runners
		err = page.Locator("#runners > a").Click()
		require.NoError(t, err)

		// expect only one runner to be listed
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-1 of 1")
		require.NoError(t, err)

		// expect otfd server to be listed as one and only runner
		err = expect.Locator(page.Locator(`#process-name`)).ToHaveText(`otfd`)
		require.NoError(t, err)

		// hide otfd runners
		err = page.Locator("input[name='hide_server_runners']").Click()
		require.NoError(t, err)

		// expect zero runners to be listed
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("0-0 of 0")
		require.NoError(t, err)
	})
}
