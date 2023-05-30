package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// TestIntegration_UserTokenUI demonstrates managing user tokens via the UI.
func TestIntegration_UserTokenUI(t *testing.T) {
	t.Parallel()

	// Create org and its owner
	svc := setup(t, nil)
	user, userCtx := svc.createUserCtx(t, ctx)
	browser := createBrowserCtx(t)
	okDialog(t, browser)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, userCtx, svc.Hostname(), user.Username, svc.Secret),
		// go to profile
		chromedp.Navigate("https://" + svc.Hostname() + "/app/profile"),
		chromedp.WaitReady(`body`),
		// go to user tokens
		chromedp.Click(`//div[@id='user-tokens-link']/a`, chromedp.NodeVisible),
		screenshot(t, "user_tokens"),
		// go to new token
		chromedp.Click(`//button[@id='new-user-token-button']`, chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// enter description for new token and submit
		chromedp.Focus("input#description", chromedp.NodeVisible),
		input.InsertText("my new token"),
		screenshot(t, "user_token_enter_description"),
		chromedp.Click(`//button[text()='Create token']`, chromedp.NodeVisible),
		screenshot(t, "user_token_created"),
		matchRegex(t, ".flash-success", `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
		// delete the token
		chromedp.Click(`//button[text()='delete']`, chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "Deleted token"),
	})
	require.NoError(t, err)
}
