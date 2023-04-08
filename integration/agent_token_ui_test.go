package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// TestAgentTokenUI demonstrates managing agent tokens via the UI.
func TestAgentTokenUI(t *testing.T) {
	t.Parallel()

	svc := setup(t, nil)
	user, ctx := svc.createUserCtx(t, ctx)
	org := svc.createOrganization(t, ctx)

	browser := createBrowserCtx(t)
	okDialog(t, browser)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, svc.Hostname(), user.Username, svc.Secret),
		chromedp.Tasks{
			// go to org main menu
			chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
			screenshot(t),
			// go to list of agent tokens
			chromedp.Click("#agent_tokens > a", chromedp.NodeVisible),
			screenshot(t),
			// go to new agent token page
			chromedp.Click(`//button[text()='New Agent Token']`, chromedp.NodeVisible),
			screenshot(t),
			// enter description for new agent token
			chromedp.Focus("input#description", chromedp.NodeVisible),
			input.InsertText("my-new-agent-token"),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[text()='Create token']`, chromedp.NodeVisible),
			screenshot(t),
			matchRegex(t, ".flash-success", `Created token: [\w-]+\.[\w-]+\.[\w-]+`),
			// delete the token
			chromedp.Click(`//button[text()='delete']`, chromedp.NodeVisible),
			screenshot(t),
			matchText(t, ".flash-success", "Deleted token: my-new-agent-token"),
		},
	})
	require.NoError(t, err)
}
