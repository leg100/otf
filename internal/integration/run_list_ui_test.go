package integration

import (
	"context"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunListUI demonstrates listing runs via the UI.
func TestIntegration_RunListUI(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	user, ctx := daemon.createUserCtx(t, ctx)
	ws := daemon.createWorkspace(t, ctx, nil)
	tfConfig := newRootModule(t, daemon.Hostname(), ws.Organization, ws.Name)

	var runListingAfter []*cdp.Node
	browser := createTabCtx(t)
	err := chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, daemon.Hostname(), user.Username, daemon.Secret),
		// navigate to workspace page
		chromedp.Navigate(workspaceURL(daemon.Hostname(), ws.Organization, ws.Name)),
		chromedp.WaitReady(`body`),
		// navigate to runs page
		chromedp.Click(`//a[text()='runs']`, chromedp.NodeVisible),
		chromedp.WaitReady(`body`),
		// should be no runs listed
		matchText(t, `//div[@id='content-list']`, `No items currently exist.`),
		chromedp.ActionFunc(func(context.Context) error {
			// meanwhile, execute a terraform cli init and plan
			daemon.tfcli(t, ctx, "init", tfConfig)
			daemon.tfcli(t, ctx, "plan", tfConfig)
			return nil
		}),
		// should be one run listed
		chromedp.Nodes(`//div[@id='content-list']//*[@class='item']`, &runListingAfter, chromedp.BySearch),
		// and its status should be 'planned and finished'
		chromedp.WaitVisible(`//*[@class='item']//*[@class='status status-planned_and_finished']`, chromedp.BySearch),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(runListingAfter))
}
