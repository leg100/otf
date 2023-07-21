package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

// TestAgentTokenUI demonstrates managing agent tokens via the UI.
func TestAgentTokenUI(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	var clipboardContent any

	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Tasks{
			// go to org main menu
			chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
			screenshot(t),
			// go to list of agent tokens
			chromedp.Click("#agent_tokens > a", chromedp.ByQuery),
			screenshot(t),
			// go to new agent token page
			chromedp.Click(`//button[text()='New Agent Token']`),
			screenshot(t),
			// enter description for new agent token
			chromedp.Focus("textarea#description", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("my-new-agent-token"),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[text()='Create token']`),
			screenshot(t),
			matchRegex(t, "//div[@role='alert']", `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
			// click clipboard icon to copy token into clipboard
			chromedp.Click(`//div[@role='alert']//img[@id='clipboard-icon']`),
			chromedp.Evaluate(`window.navigator.clipboard.readText()`, &clipboardContent, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
				return p.WithAwaitPromise(true)
			}),
			// delete the token
			chromedp.Click(`//button[text()='delete']`),
			screenshot(t),
			matchText(t, `//div[@role='alert']`, `Deleted token: my-new-agent-token`),
		},
	})

	// clipboard should contained agent token (base64 encoded JWT) and no white
	// space.
	assert.Regexp(t, `^[\w-]+\.[\w-]+\.[\w-]+$`, clipboardContent)
}
