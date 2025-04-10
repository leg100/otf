package integration

import (
	"regexp"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_UserTokenUI demonstrates managing user tokens via the UI.
func TestIntegration_UserTokenUI(t *testing.T) {
	integrationTest(t)

	svc, _, ctx := setup(t, nil)
	browser.New(t, ctx, func(page playwright.Page) {
		// go to profile
		_, err := page.Goto("https://" + svc.System.Hostname() + "/app/profile")
		require.NoError(t, err)

		// go to user tokens
		err = page.Locator(`#menu-item-user-tokens > a`).Click()
		require.NoError(t, err)

		screenshot(t, page, "user_tokens")
		// go to new token
		err = page.Locator(`//button[@id='new-user-token-button']`).Click()
		require.NoError(t, err)

		// enter description for new token and submit
		err = page.Locator("textarea#description").Fill("my new token")
		require.NoError(t, err)
		screenshot(t, page, "user_token_enter_description")

		err = page.Locator(`//button[text()='Create token']`).Click()
		require.NoError(t, err)

		screenshot(t, page, "user_token_created")
		err = expect.Locator(page.GetByRole("alert")).ToHaveText(regexp.MustCompile(`Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`))
		require.NoError(t, err)

		// delete the token
		err = page.Locator(`//form[@id='delete-user-token']/button`).Click()
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText("Deleted token")
		require.NoError(t, err)
	})
}
