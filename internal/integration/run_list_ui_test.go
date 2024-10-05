package integration

import (
	"context"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_RunListUI demonstrates listing runs via the UI.
func TestIntegration_RunListUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, nil)
	ws := daemon.createWorkspace(t, ctx, nil)
	tfConfig := newRootModule(t, daemon.System.Hostname(), ws.Organization, ws.Name)

	var runListingAfter []*cdp.Node
	page := browser.New(t, ctx)
		// navigate to workspace page
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), ws.Organization, ws.Name))
require.NoError(t, err)
		// navigate to runs page
		err := page.Locator(`//a[text()='runs']`).Click()
require.NoError(t, err)
		// should be no runs listed
		matchText(t, `//div[@id='content-list']`, `No items currently exist.`),
		chromedp.ActionFunc(func(context.Context) error {
			// meanwhile, execute a terraform cli init and plan
			daemon.tfcli(t, ctx, "init", tfConfig)
			daemon.tfcli(t, ctx, "plan", tfConfig)
			return nil
		}),
		// should be one run listed
		chromedp.Nodes(`//div[@id='content-list']//*[@class='widget']`, &runListingAfter, chromedp.NodeVisible),
		// and its status should be 'planned and finished'
		chromedp.WaitVisible(`//*[@class='widget']//a[text()='planned and finished']`),
	})
	assert.Equal(t, 1, len(runListingAfter))
}
