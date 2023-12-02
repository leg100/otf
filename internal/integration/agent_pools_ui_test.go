package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

// TestAgentPoolsUI demonstrates managing agent pools and tokens via the UI.
func TestAgentPoolsUI(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	var clipboardContent any

	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Tasks{
			// go to org main menu
			chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
			// go to list of agent pools
			chromedp.Click("#agent_pools > a", chromedp.ByQuery),
			screenshot(t),
			// expose new agent pool form
			chromedp.Click("#new-pool-details", chromedp.ByQuery),
			screenshot(t),
			// enter name for new agent pool
			chromedp.Focus("input#new-pool-name", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("pool-1"),
			// submit form
			chromedp.Click(`//button[text()='Create agent pool']`),
			screenshot(t),
			// expect flash message confirming pool creation
			matchText(t, `//div[@role='alert']`, `created agent pool: pool-1`),
			// expose new agent token form
			chromedp.Click("#new-token-details", chromedp.ByQuery),
			// enter description for new agent token
			chromedp.Focus("input#new-token-description", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("token-1"),
			// submit form
			chromedp.Click(`//button[text()='Create token']`),
			screenshot(t),
			// expect flash message confirming token creation
			matchRegex(t, `//div[@role='alert']`, `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
			// click clipboard icon to copy token into clipboard
			chromedp.Click(`//div[@role='alert']//img[@id='clipboard-icon']`),
			chromedp.Evaluate(`window.navigator.clipboard.readText()`, &clipboardContent, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
				return p.WithAwaitPromise(true)
			}),
			// delete the token
			chromedp.Click(`//button[@id="delete-agent-token-button"]`),
			screenshot(t),
			matchText(t, `//div[@role='alert']`, `Deleted token: token-1`),
		},
	})

	// clipboard should contained agent token (base64 encoded JWT) and no white
	// space.
	assert.Regexp(t, `^[\w-]+\.[\w-]+\.[\w-]+$`, clipboardContent)
}
