package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

// TestIntegration_UserTokenUI demonstrates managing user tokens via the UI.
func TestIntegration_UserTokenUI(t *testing.T) {
	integrationTest(t)

	svc, _, ctx := setup(t, nil)
	page := page := browser.New(t, ctx)
		// go to profile
		_, err = page.Goto("https://" + svc.System.Hostname() + "/app/profile")
require.NoError(t, err)
		// go to user tokens
		err := page.Locator(`//div[@id='user-tokens-link']/a`).Click()
require.NoError(t, err)
		////screenshot(t, "user_tokens"),
		// go to new token
		err := page.Locator(`//button[@id='new-user-token-button']`).Click()
require.NoError(t, err)
		// enter description for new token and submit
		chromedp.Focus("textarea#description", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("my new token"),
		//screenshot(t, "user_token_enter_description"),
		err := page.Locator(`//button[text()='Create token']`).Click()
require.NoError(t, err)
		//screenshot(t, "user_token_created"),
		matchRegex(t, "//div[@role='alert']", `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
		// delete the token
		err := page.Locator(`//button[text()='delete']`).Click()
require.NoError(t, err)
		//screenshot(t),
		matchText(t, "//div[@role='alert']", "Deleted token"),
	})
}
