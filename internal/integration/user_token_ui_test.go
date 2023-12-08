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
	browser.Run(t, ctx, chromedp.Tasks{
		// go to profile
		chromedp.Navigate("https://" + svc.System.Hostname() + "/app/profile"),
		// go to user tokens
		chromedp.Click(`//div[@id='user-tokens-link']/a`),
		screenshot(t, "user_tokens"),
		// go to new token
		chromedp.Click(`//button[@id='new-user-token-button']`),
		// enter description for new token and submit
		chromedp.Focus("textarea#description", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("my new token"),
		screenshot(t, "user_token_enter_description"),
		chromedp.Click(`//button[text()='Create token']`),
		screenshot(t, "user_token_created"),
		matchRegex(t, "//div[@role='alert']", `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
		// delete the token
		chromedp.Click(`//button[text()='delete']`),
		screenshot(t),
		matchText(t, "//div[@role='alert']", "Deleted token"),
	})
}
