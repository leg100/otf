package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentTokenUI demonstrates managing agent tokens via the UI.
func TestAgentTokenUI(t *testing.T) {
	t.Parallel()

	svc, org, _ := setup(t, nil)

	var clipboardContent any

	tab := createTabCtx(t)
	err := chromedp.Run(tab, chromedp.Tasks{
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
			matchRegex(t, ".flash-success", `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
			// click clipboard icon to copy token into clipboard
			chromedp.Click(`//div[@class='flash flash-success']//img[@class='clipboard-icon']`, chromedp.BySearch),
			chromedp.Evaluate(`window.navigator.clipboard.readText()`, &clipboardContent, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
				return p.WithAwaitPromise(true)
			}),
			// delete the token
			chromedp.Click(`//button[text()='delete']`, chromedp.NodeVisible),
			screenshot(t),
			matchText(t, ".flash-success", "Deleted token: my-new-agent-token"),
		},
	})
	require.NoError(t, err)

	// clipboard should contained agent token (base64 encoded JWT) and no white
	// space.
	assert.Regexp(t, `^[\w-]+\.[\w-]+\.[\w-]+$`, clipboardContent)
}
