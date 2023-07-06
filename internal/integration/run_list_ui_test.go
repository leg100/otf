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
	tfConfig := newRootModule(t, daemon.Hostname(), ws.Organization, ws.Name)

	var runListingAfter []*cdp.Node
	browser.Run(t, ctx, chromedp.Tasks{
		// navigate to workspace page
		chromedp.Navigate(workspaceURL(daemon.Hostname(), ws.Organization, ws.Name)),
		// navigate to runs page
		chromedp.Click(`//a[text()='runs']`),
		// should be no runs listed
		matchText(t, `//div[@id='content-list']`, `No items currently exist.`),
		chromedp.ActionFunc(func(context.Context) error {
			// meanwhile, execute a terraform cli init and plan
			daemon.tfcli(t, ctx, "init", tfConfig)
			daemon.tfcli(t, ctx, "plan", tfConfig)
			return nil
		}),
		// should be one run listed
		chromedp.Nodes(`//div[@id='content-list']//*[@class='item']`, &runListingAfter, chromedp.NodeVisible),
		// and its status should be 'planned and finished'
		chromedp.WaitVisible(`//*[@class='item']//*[@class='status status-planned_and_finished']`),
	})
	assert.Equal(t, 1, len(runListingAfter))
}
